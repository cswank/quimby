#!/bin/bash

export QUIMBY_STATIC="www/dist"
cd www
rm -rf dist
bower install
grunt
cd ..
rice embed-go
