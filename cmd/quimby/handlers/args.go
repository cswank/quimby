package handlers

import (
	"net/url"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby"
)

type Controller func(args *Args) error
type caller func()

type Args struct {
	DB     *bolt.DB
	User   *quimby.User
	Gadget *quimby.Gadget
	Vars   map[string]string
	Args   url.Values
	LG     quimby.Logger
}
