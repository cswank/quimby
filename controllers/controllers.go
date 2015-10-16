package controllers

import (
	"os"

	"github.com/boltdb/bolt"
	"github.com/gorilla/securecookie"
)

type Logger interface {
	Printf(string, ...interface{})
	Println(...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

var (
	DB       *bolt.DB
	addr     string
	Clients  *ClientHolder
	host     string
	LG       Logger
	hashKey  = []byte(os.Getenv("QUIMBY_HASH_KEY"))
	blockKey = []byte(os.Getenv("QUIMBY_BLOCK_KEY"))
	sc       = securecookie.New(hashKey, blockKey)
)
