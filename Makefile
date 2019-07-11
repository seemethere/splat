bin/splat: splat.go
	go build -o $@ .

clean:
	$(RM) -r bin

