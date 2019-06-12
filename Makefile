bin/splat: main.go
	go build -o $@ .

clean:
	$(RM) -r bin

