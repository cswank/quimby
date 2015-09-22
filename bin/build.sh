#!/bin/bash

export QUIMBY_STATIC="www/dist"
cd www
rm -rf dist
npm install
./node_modules/.bin/grunt
cd ..
rice embed-go
godep go install
