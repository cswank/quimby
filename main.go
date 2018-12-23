package main

import (
	"net/http"

	userhttp "github.com/cswank/quimby/internal/user/delivery/http"
	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()
	userhttp.New(r)
	http.ListenAndServe(":3333", r)
}
