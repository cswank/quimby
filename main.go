package main

import (
	"fmt"
	"log"

	"github.com/cswank/quimby/internal/auth"
	"github.com/cswank/quimby/internal/config"
	"github.com/cswank/quimby/internal/homekit"
	"github.com/cswank/quimby/internal/repository"
	"github.com/cswank/quimby/internal/router"
	"github.com/cswank/quimby/internal/user"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	srv = kingpin.Command("serve", "start server")

	gdt    = kingpin.Command("gadget", "gadget crud")
	create = gdt.Command("create", "create a gadget")
	name   = create.Flag("name", "name of the gadget").Short('n').Required().String()
	url    = create.Flag("url", "url of the gadget").Short('u').Required().String()
	del    = gdt.Command("delete", "delete a gadget")
	id     = del.Arg("id", "id of the gadget").Required().Int()

	usr      = kingpin.Command("user", "user crud")
	mkUser   = usr.Command("create", "create a user")
	username = mkUser.Arg("username", "username").Required().String()

	delUser     = usr.Command("delete", "delete a user")
	delUsername = delUser.Arg("username", "username").Required().String()

	_ = usr.Command("list", "list users")
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
		serve(cfg, g, u)
	case "user create":
		createUser(u)
	case "user delete":
		deleteUser(u)
	case "user list":
		listUsers(u)
	case "gadget create":
		createGadget(g)
	case "gadget delete":
		deleteGadget(g)
	case "gadget update":
		updateGadget(g)
	}
}

func serve(cfg config.Config, g *repository.Gadget, u *repository.User) {
	a := auth.New(u)
	hc, err := homekit.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := router.Serve(cfg, g, u, a, hc); err != nil {
		log.Fatal(err)
	}
}

func createGadget(r *repository.Gadget) {
	g, err := r.Create(*name, *url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created gadget\n: %+v\n", g)
}

func deleteGadget(r *repository.Gadget) {
	if err := r.Delete(*id); err != nil {
		log.Fatal(err)
	}
}

func updateGadget(r *repository.Gadget) {
	// if err := r.Update(*id); err != nil {
	// 	log.Fatal(err)
	// }
}

func listGadgets(r *repository.Gadget) {
	gds, err := r.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, g := range gds {
		fmt.Printf("%d: %s %s\n", g.ID, g.Name, g.URL)
	}
}

func editGadget(r *repository.Gadget) {
	// if err := r.Edit(*id, *username, *url); err != nil {
	// 	log.Fatal(err)
	// }
}

func createUser(r *repository.User) {
	if err := user.Create(r, *username); err != nil {
		log.Fatal(err)
	}
}

func deleteUser(r *repository.User) {
	if err := r.Delete(*delUsername); err != nil {
		log.Fatal(err)
	}
}

func listUsers(r *repository.User) {
	users, err := r.GetAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		fmt.Printf("%d: %s\n", u.ID, u.Name)
	}
}
