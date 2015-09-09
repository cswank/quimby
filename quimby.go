package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var (
	hashKey      []byte
	blockKey     []byte
	static       string
	SecureCookie *securecookie.SecureCookie
)

func init() {
	hashKey = []byte(os.Getenv("GADGETS_HASH_KEY"))
	blockKey = []byte(os.Getenv("GADGETS_BLOCK_KEY"))
	static = os.Getenv("GADGETS_STATIC")
	if len(static) == 0 {
		log.Fatal("you must set the GADGETS_STATIC env var")
	}
	SecureCookie = securecookie.New(hashKey, blockKey)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/api/login", auth.Login).Methods("POST")
	r.HandleFunc("/api/logout", auth.Logout).Methods("POST")
	r.HandleFunc("/api/gadgets", GetGadgets).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(static)))

	http.Handle("/", r)
	fmt.Println("listening on 0.0.0.0:8080", static)
	err := http.ListenAndServe(":8080", r)
	log.Println(err)
}

func GetGadgets(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.GetGadgets, "read")
}
