module github.com/seemethere/splat

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.8.1 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli v1.20.0
	golang.org/x/net v0.0.0-20190603091049-60506f45cf65 // indirect
	golang.org/x/xerrors v0.0.0-20190513163551-3ee3066db522
)

replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190109173153-a79fabbfe841
