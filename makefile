all:
	rice embed-go
linux: all
	GOOS=linux GOARCH=amd64 go build .
	rm *rice-box.go
	echo 'done'
macos: all
	GOOS=darwin GOARCH=amd64 go build .
	rm *rice-box.go
	echo 'done'
