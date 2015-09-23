package controllers

import (
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
)

type Logger interface {
	Printf(string, ...interface{})
	Println(...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

type Args struct {
	W      http.ResponseWriter
	R      *http.Request
	DB     *bolt.DB
	User   *models.User
	Gadget *models.Gadget
	Vars   map[string]string
	LG     Logger
}
