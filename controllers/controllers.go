package controllers

import (
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby/models"
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
	clients  map[string]chan gogadgets.Message
	host     string
	hashKey  = []byte(os.Getenv("QUIMBY_HASH_KEY"))
	blockKey = []byte(os.Getenv("QUIMBY_BLOCK_KEY"))
	sc       = securecookie.New(hashKey, blockKey)
)

type Args struct {
	W      http.ResponseWriter
	R      *http.Request
	DB     *bolt.DB
	User   *models.User
	Gadget *models.Gadget
	Vars   map[string]string
	LG     Logger
}
