package main

import (
	"log"

	"github.com/cswank/quimby/internal/commandline/gadget"
	"github.com/cswank/quimby/internal/commandline/server"
	"github.com/cswank/quimby/internal/commandline/user"
	"github.com/cswank/quimby/internal/config"
	"github.com/cswank/quimby/internal/repository"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	srv = kingpin.Command("serve", "start server")

	// gadget commands
	gdt    = kingpin.Command("gadget", "gadget crud")
	create = gdt.Command("create", "create a gadget")
	name   = create.Flag("name", "name of the gadget").Short('n').Required().String()
	url    = create.Flag("url", "url of the gadget").Short('u').Required().String()
	del    = gdt.Command("delete", "delete a gadget")
	did    = del.Arg("id", "id of the gadget").Required().Int()
	up     = gdt.Command("update", "update a gadget")
	uname  = up.Flag("name", "name of the gadget").Short('n').Required().String()
	uurl   = up.Flag("url", "url of the gadget").Short('u').Required().String()
	uid    = up.Arg("id", "id of the gadget").Required().Int()
	_      = gdt.Command("list", "list gadgets")

	// user commands
	usr       = kingpin.Command("user", "user crud")
	mkUser    = usr.Command("create", "create a user")
	username  = mkUser.Arg("username", "username").Required().String()
	duser     = usr.Command("delete", "delete a user")
	dusername = duser.Arg("username", "username").Required().String()
	_         = usr.Command("list", "list users")
)

func main() {
	cfg := config.Get()
	g, u, err := repository.New(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}

	cmd := kingpin.Parse()
	switch cmd {
	case "serve":
		server.Start(cfg, g, u)
	case "user create":
		user.Create(u, *username)
	case "user delete":
		user.Delete(u, *dusername)
	case "user list":
		user.List(u)
	case "gadget create":
		gadget.Create(g, *name, *url)
	case "gadget delete":
		gadget.Delete(g, *did)
	case "gadget update":
		gadget.Update(g, *uid, *uname, *uurl)
	case "gadget list":
		gadget.List(g)
	}
}
