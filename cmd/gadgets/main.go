package main

import (
	"fmt"
	"log"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/cswank/quimby/internal/gadget"
	"github.com/cswank/quimby/internal/gadget/repository"
	"github.com/cswank/quimby/internal/storage"
)

var (
	gdt = kingpin.Command("gadget", "gadget crud")

	create = gdt.Command("create", "create a gadget")
	name   = create.Flag("name", "name of the gadget").Short('n').Required().String()
	url    = create.Flag("url", "url of the gadget").Short('u').Required().String()
)

func main() {
	defer storage.Close()

	repo := repository.New()
	cmd := kingpin.Parse()
	var err error
	switch cmd {
	case "gadget create":
		err = doCreate(repo, *name, *url)
	default:
		log.Printf("unknown command '%s'", cmd)
	}

	if err != nil {
		log.Println("oops, something went wrong: %s", err)
	}
}

func doCreate(repo gadget.Repository, name, url string) error {
	g, err := repo.Create(name, url)
	fmt.Printf("%+v\n", g)
	return err
}
