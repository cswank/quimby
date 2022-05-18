package main

import (
	"fmt"
	"log"

	"github.com/cswank/quimby/internal/auth"
	"github.com/cswank/quimby/internal/config"
	"github.com/cswank/quimby/internal/homekit"
	"github.com/cswank/quimby/internal/repository"
	"github.com/cswank/quimby/internal/router"
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
	ed     = gdt.Command("edit", "edit a gadget")
	edID   = ed.Arg("id", "id of the gadget").Required().Int()
	ls     = gdt.Command("ls", "list gadgets")

	usr       = kingpin.Command("user", "user crud")
	createUsr = usr.Command("create", "create a user")
	username  = createUsr.Arg("username", "username").Required().String()

	delUsr      = usr.Command("delete", "delete a user")
	delUsername = delUsr.Arg("username", "username").Required().String()
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
		serve(g, u)
	}
}

func serve(g *repository.Gadget, u *repository.User) {
	a := auth.New(u)
	hc, err := homekit.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := router.Serve(g, u, a, hc); err != nil {
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

func listGadgets(r *repository.Gadget) {
	gds, err := r.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, g := range gds {
		fmt.Printf("%+v\n", g)
	}
}

func editGadget(r *repository.Gadget) {
	// if err := r.Edit(*id, *username, *url); err != nil {
	// 	log.Fatal(err)
	// }
}

func createUser(r *repository.User) {
	// fmt.Print("Enter Password: ")
	// pw, err := terminal.ReadPassword(int(syscall.Stdin))
	// if err != nil {
	// 	return err
	// }

	// uc := userusecase.New()
	// u, qa, err := uc.Create(name, string(pw))
	// if err != nil {
	// 	return err
	// }

	// f, err := ioutil.TempFile("", "")
	// if err != nil {
	// 	return err
	// }

	// _, err = io.Copy(f, bytes.NewBuffer(qa))
	// if err != nil {
	// 	return err
	// }

	// fmt.Printf("created user\n: %+v, scan qa code at %s (and then delete it)\n", u, f.Name())
	// return nil
}

func deleteUser(r *repository.User) {
	// uc := userusecase.New()
	// return uc.Delete(name)
}
