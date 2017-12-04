package handlers

import (
	"net/url"

	"github.com/cswank/quimby"
)

type Controller func(args *Args) error
type caller func()

type Args struct {
	User   *quimby.User
	Gadget *quimby.Gadget
	Vars   map[string]string
	Args   url.Values
	LG     quimby.Logger
}
