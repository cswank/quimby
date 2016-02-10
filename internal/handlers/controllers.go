package handlers

import (
	"github.com/boltdb/bolt"
	"github.com/cswank/quimby"
)

var (
	DB *bolt.DB
	LG quimby.Logger
)
