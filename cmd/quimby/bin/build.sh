#!/bin/bash

rice embed-go
go build
rm *rice-box.go
echo 'done'
