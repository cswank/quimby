package main

import (
	"log"
	"net/http"

	"bitbucket.org/cswank/rtree"
)

func main() {
	r := rtree.New()
	r.Get("/api/login", Login)

	http.Handle("/", r)
	log.Println(http.ListenAndServe(":8888", r))
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}
