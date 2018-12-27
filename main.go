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

func doCreate(name, url string) error {
	repo := repository.New()
	g, err := repo.Create(name, url)
	fmt.Printf("%+v\n", g)
	return err
}

func doServe() error {
	box := rice.MustFindBox("templates")
	templates.Box(box)
	r := chi.NewRouter()

	userhttp.New(r)
	gadgethttp.New(r, box)
	server := getServer(r)
	return server.ListenAndServeTLS("server_cert.pem", "server_key.pem") //private cert
}

func getServer(h http.Handler) *http.Server {
	caCert, err := ioutil.ReadFile("server_cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	return &http.Server{
		Addr:      ":3333",
		TLSConfig: tlsConfig,
		Handler:   h,
	}
}
