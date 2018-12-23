package main

import (
	"net/http"

	gadgethttp "github.com/cswank/quimby/internal/gadget/delivery/http"
	userhttp "github.com/cswank/quimby/internal/user/delivery/http"
	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()
	userhttp.New(r)
	gadgethttp.New(r)
	http.ListenAndServe(":3333", r)
}
