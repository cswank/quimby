#!/bin/bash

cd www
rm -rf dist
npm install
./node_modules/.bin/grunt
cd ..
rice embed-go
godep go install
rm -rf dist
cd www
ln -s app dist
cd ..

