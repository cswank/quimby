package controllers

import (
	"os"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/securecookie"
)

var (
	DB       *bolt.DB
	LG       models.Logger
	hashKey  = []byte(os.Getenv("QUIMBY_HASH_KEY"))
	blockKey = []byte(os.Getenv("QUIMBY_BLOCK_KEY"))
	sc       = securecookie.New(hashKey, blockKey)
)
