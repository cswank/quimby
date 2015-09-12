package controllers

import (
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
)

type Args struct {
	W      http.ResponseWriter
	R      *http.Request
	DB     *bolt.DB
	User   *models.User
	Gadget *models.Gadget
	Vars   map[string]string
}
