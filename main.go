package main

import (
	"log"
	"net/http"

	gadgethttp "github.com/cswank/quimby/internal/gadget/delivery/http"
	userhttp "github.com/cswank/quimby/internal/user/delivery/http"
	"github.com/go-chi/chi"
	"github.com/gobuffalo/packr"
)

func main() {
	r := chi.NewRouter()

	box := packr.NewBox("./html")
	userhttp.New(r)
	gadgethttp.New(r)
	if err := http.ListenAndServe(":3333", r); err != nil {
		log.Fatal(err)
	}
}
