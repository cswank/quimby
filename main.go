package main

import (
	"log"
	"net/http"

	gadgethttp "github.com/cswank/quimby/internal/gadget/delivery/http"
	"github.com/cswank/quimby/internal/storage"
	userhttp "github.com/cswank/quimby/internal/user/delivery/http"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()

	defer storage.Close()

	userhttp.New(r)
	gadgethttp.New(r)
	if err := http.ListenAndServe(":3333", r); err != nil {
		log.Fatal(err)
	}
}
