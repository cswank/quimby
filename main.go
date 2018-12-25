package main

import (
	"fmt"
	"log"
	"net/http"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	rice "github.com/GeertJohan/go.rice"
	gadgethttp "github.com/cswank/quimby/internal/gadget/delivery/http"
	"github.com/cswank/quimby/internal/gadget/repository"
	"github.com/cswank/quimby/internal/storage"
	"github.com/cswank/quimby/internal/templates"
	userhttp "github.com/cswank/quimby/internal/user/delivery/http"

	"github.com/go-chi/chi"
)

var (
	_ = kingpin.Command("serve", "start server")

	gdt    = kingpin.Command("gadget", "gadget crud")
	create = gdt.Command("create", "create a gadget")
	name   = create.Flag("name", "name of the gadget").Short('n').Required().String()
	url    = create.Flag("url", "url of the gadget").Short('u').Required().String()
)

func main() {
	defer storage.Close()

	var err error
	cmd := kingpin.Parse()
	switch cmd {
	case "serve":
		err = doServe()
	case "gadget create":
		err = doCreate(*name, *url)
	default:
		log.Printf("unknown command '%s'", cmd)
	}

	if err != nil {
		log.Println("oops, something went wrong: %s", err)
	}

}

func doServe() error {
	box := rice.MustFindBox("html")
	templates.Box(box)
	r := chi.NewRouter()

	userhttp.New(r)
	gadgethttp.New(r)
	return http.ListenAndServe(":3333", r)
}

func doCreate(name, url string) error {
	repo := repository.New()
	g, err := repo.Create(name, url)
	fmt.Printf("%+v\n", g)
	return err
}
