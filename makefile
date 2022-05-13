all:
	rice embed-go
linux: all
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -tags netgo . 
	rm *rice-box.go
macos: all
	GOOS=darwin GOARCH=amd64 go build .
	rm *rice-box.go
