all:
	rice embed-go
	GOOS=linux GOARCH=amd64 go build .
	rm *rice-box.go
	echo 'done'
