#!/bin/bash

cd www
rm -rf dist
npm install
./node_modules/.bin/grunt

cd ..
rice embed-go
go build
rm www-dist.rice-box.go
cd www
rm -r dist
ln -s app dist
cd ..
