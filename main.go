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
	"github.com/cswank/quimby/internal/config"
	gadgethttp "github.com/cswank/quimby/internal/gadget/delivery/http"
	"github.com/cswank/quimby/internal/gadget/repository"
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

	usr        = kingpin.Command("user", "user crud")
	createUser = usr.Command("create", "create a user")
	username   = createUser.Arg("username", "username").Required().String()
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
	case "user create":
		err = doCreateUser(*username)
	default:
		err = fmt.Errorf("unknown command '%s'", cmd)
	}

	if err != nil {
		log.Println("oops, something went wrong: %s", err)
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

	fmt.Printf("created user\n: %+v, scan qa code at %s (and then delete it), \n", u, f.Name())
	return nil
}

func doServe() error {
	box := rice.MustFindBox("templates")
	templates.Box(box)

	pub := chi.NewRouter()
	priv := chi.NewRouter()
	gadgethttp.Init(pub, priv, box)
	userhttp.Init(pub, box)

	go func(r chi.Router) {
		if err := http.ListenAndServe(":3334", r); err != nil {
			log.Fatalf("unable to start private server: %s", err)
		}
	}(priv)

	cfg := config.Get()
	return http.ListenAndServeTLS(":3333", cfg.TLSCert, cfg.TLSKey, pub)
}
