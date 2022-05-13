package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"syscall"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	rice "github.com/GeertJohan/go.rice"
	gadgethttp "github.com/cswank/quimby/internal/gadget/delivery/http"
	"github.com/cswank/quimby/internal/gadget/repository"
	"github.com/cswank/quimby/internal/homekit"
	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/storage"
	"github.com/cswank/quimby/internal/templates"
	userhttp "github.com/cswank/quimby/internal/user/delivery/http"
	userusecase "github.com/cswank/quimby/internal/user/usecase"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/go-chi/chi"
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
	edName = ed.Flag("name", "name of the gadget").Short('n').Required().String()
	edUrl  = ed.Flag("url", "url of the gadget").Short('u').Required().String()
	ls     = gdt.Command("list", "list gadgets")

	usr        = kingpin.Command("user", "user crud")
	createUser = usr.Command("create", "create a user")
	username   = createUser.Arg("username", "username").Required().String()

	delUser     = usr.Command("delete", "delete a user")
	delUsername = delUser.Arg("username", "username").Required().String()
)

func main() {
	defer storage.Close()

	var err error
	cmd := kingpin.Parse()
	switch cmd {
	case "serve":
		err = doServe()
	case "gadget create":
		err = doCreateGadget(*name, *url)
	case "gadget delete":
		err = doDeleteGadget(*id)
	case "gadget list":
		err = doListGadgets()
	case "gadget edit":
		err = doEditGadget(*edID, *edName, *edUrl)
	case "user create":
		err = doCreateUser(*username)
	case "user delete":
		err = doDeleteUser(*delUsername)
	default:
		err = fmt.Errorf("unknown command '%s'", cmd)
	}

	if err != nil {
		log.Printf("oops, something went wrong: %s", err)
	}

}

func doCreateGadget(name, url string) error {
	repo := repository.New()
	g, err := repo.Create(name, url)
	if err != nil {
		return err
	}

	fmt.Printf("created gadget\n: %+v\n", g)
	return nil
}

func doDeleteGadget(id int) error {
	repo := repository.New()
	return repo.Delete(id)
}

func doListGadgets() error {
	repo := repository.New()
	gds, err := repo.GetAll()
	if err != nil {
		return err
	}
	for _, g := range gds {
		fmt.Printf("%+v\n", g)
	}

	return nil
}

func doEditGadget(id int, name, url string) error {
	repo := repository.New()
	return repo.Edit(schema.Gadget{ID: id, Name: name, URL: url})
}

func doCreateUser(name string) error {
	fmt.Print("Enter Password: ")
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	uc := userusecase.New()
	u, qa, err := uc.Create(name, string(pw))
	if err != nil {
		return err
	}

	f, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}

	_, err = io.Copy(f, bytes.NewBuffer(qa))
	if err != nil {
		return err
	}

	fmt.Printf("created user\n: %+v, scan qa code at %s (and then delete it)\n", u, f.Name())
	return nil
}

func doDeleteUser(name string) error {
	uc := userusecase.New()
	return uc.Delete(name)
}

func doServe() error {
	hc, err := homekit.New()
	if err != nil {
		log.Fatal(err)
	}

	box := rice.MustFindBox("templates")
	templates.Box(box)

	pub := chi.NewRouter()
	priv := chi.NewRouter()

	gadgethttp.Handle(pub, priv, box, hc)
	userhttp.Handle(pub, box)

	go func(r chi.Router) {
		if err := http.ListenAndServe(":3334", r); err != nil {
			log.Fatalf("unable to start private server: %s", err)
		}
	}(priv)

	return http.ListenAndServe(":3333", pub)
}
