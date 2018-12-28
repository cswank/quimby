package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	rice "github.com/GeertJohan/go.rice"
	gadgethttp "github.com/cswank/quimby/internal/gadget/delivery/http"
	"github.com/cswank/quimby/internal/gadget/repository"
	"github.com/cswank/quimby/internal/storage"
	"github.com/cswank/quimby/internal/templates"

	"github.com/go-chi/chi"
)

var (
	srv  = kingpin.Command("serve", "start server")
	cert = srv.Flag("cert", "tls certificate file").Short('c').Required().String()
	key  = srv.Flag("key", "tls key file").Short('k').Required().String()

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
		err = fmt.Errorf("unknown command '%s'", cmd)
	}

	if err != nil {
		log.Println("oops, something went wrong: %s", err)
	}

}

func doCreate(name, url string) error {
	repo := repository.New()
	g, err := repo.Create(name, url)
	if err != nil {
		return err
	}

	fmt.Printf("created gadget\n: %+v\n", g)
	return nil
}

func doServe() error {
	box := rice.MustFindBox("templates")
	templates.Box(box)

	pub := chi.NewRouter()
	priv := chi.NewRouter()
	gadgethttp.Init(pub, priv, box)

	go func(r chi.Router) {
		if err := http.ListenAndServe(":3334", r); err != nil {
			log.Fatalf("unable to start private server: %s", err)
		}
	}(priv)

	server := getServer(pub)
	return server.ListenAndServeTLS(*cert, *key)
}

func getServer(h http.Handler) *http.Server {
	caCert, err := ioutil.ReadFile(*cert)
	if err != nil {
		log.Fatal(err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)

	cfg := &tls.Config{
		ClientCAs:  pool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	cfg.BuildNameToCertificate()

	return &http.Server{
		Addr:      ":3333",
		TLSConfig: cfg,
		Handler:   h,
	}
}
