package main

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	tarFile = regexp.MustCompile("^.*.tar")
)

type Manifest struct {
	Layers []string `json:"layers"`
}

func doSplat(c *cli.Context) error {
	ctx := context.TODO()
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	docker.NegotiateAPIVersion(ctx)

	switch {
	case c.NArg() == 0:
		return fmt.Errorf("No args specfied, see --help for usage")
	case c.NArg() > 2:
		return fmt.Errorf("Too many args, see --help for usage")
	}

	// TODO maybe this can support a list of things instead of just one eventually
	ourImage := c.Args().Get(0)
	outDir := c.Args().Get(1)
	log.Debugf("Loading image %s", ourImage)
	save, err := fetchImage(docker, ourImage)
	if err != nil {
		return err
	}
	defer save.Close()
	tarReader := tar.NewReader(save)
	var manifests []Manifest
	layerTars := make(map[string][]byte)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch {
		case hdr.Name == "manifest.json":
			byteVal, err := ioutil.ReadAll(tarReader)
			if err != nil {
				return err
			}
			err = json.Unmarshal(byteVal, &manifests)
			if err != nil {
				return err
			}
		case tarFile.MatchString(hdr.Name):
			layerTars[hdr.Name], err = ioutil.ReadAll(tarReader)
			if err != nil {
				return err
			}
		}
	}
	for _, manifest := range manifests {
		for _, layer := range manifest.Layers {
			tarBytes, found := layerTars[layer]
			if !found {
				panic("wtf")
			}
			log.Debugf("Unpacking '%s' size: %d", layer, len(tarBytes))
			err := unpackLayer(tarBytes, outDir)
			if err != nil {
				return err
			}
			// TODO: Add thing to iterate through tar layer and apply to file system
		}
	}
	return nil
}

func unpackLayer(tarBytes []byte, outDir string) error {
	tarReader := tar.NewReader(bytes.NewReader(tarBytes))
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		base := filepath.Base(hdr.Name)
		dir := filepath.Dir(hdr.Name)
		outFile := filepath.Join(outDir, hdr.Name)
		switch {
		// whiteout file, delete files
		case strings.HasPrefix(base, ".wh."):
			outFile := filepath.Join("out", dir, base[4:])
			log.Debugf("whiteout detected, removing %s", outFile)
			fileInfo, err := os.Stat(outFile)
			if err != nil {
				log.Warnf("could not find file %s", outFile)
				continue
			}
			remove := os.Remove
			if fileInfo.IsDir() {
				remove = os.RemoveAll
			}

			if err = remove(outFile); err != nil {
				return err
			}
		// Taken from https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
		case hdr.Typeflag == tar.TypeDir:
			if _, err := os.Stat(outFile); err != nil {
				log.Debugf("creating directory %s", outFile)
				if err := os.MkdirAll(outFile, 0755); err != nil {
					return err
				}
			}
		case hdr.Typeflag == tar.TypeReg:
			f, err := os.OpenFile(outFile, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}

			log.Debugf("creating file %s", outFile)
			// copy over contents
			if _, err := io.Copy(f, tarReader); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
	return nil
}

func fetchImage(docker client.APIClient, ref string) (io.ReadCloser, error) {
	ctx := context.TODO()
	save, err := docker.ImageSave(ctx, []string{ref})
	if err != nil {
		log.Infof("Image not found, attempting to pull it...")
		// Attempt to pull image before declaring bankruptcy
		_, err = docker.ImagePull(ctx, ref, types.ImagePullOptions{All: true})
		if err != nil {
			return nil, err
		}
		save, err = docker.ImageSave(ctx, []string{ref})
		if err != nil {
			return nil, err
		}
	}
	return save, nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "take container images and put them on your filesystem"
	app.UsageText = fmt.Sprintf("%s [container image]:source] [destination]", app.Name)
	app.Action = doSplat
	app.Version = "0.0.1"
	log.SetLevel(log.DebugLevel)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
