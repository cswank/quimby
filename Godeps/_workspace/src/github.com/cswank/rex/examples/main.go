package main

import (
	"log"
	"net/http"

	"github.com/cswank/rex"
)

func main() {
	r := rex.New("example")
	r.Get("/api/login", Login)
	r.ServeFiles(http.FileServer(http.Dir("./www")))

	http.Handle("/", r)

	log.Println(http.ListenAndServe(":8888", r))
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}
