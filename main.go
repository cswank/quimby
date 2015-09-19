package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/GeertJohan/go.rice"
	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/auth"
	"github.com/cswank/quimby/controllers"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/mux"
	"github.com/howeyc/gopass"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	static   string
	dbPath   string
	app      = kingpin.New("quimby", "An interface to gogadets")
	register = app.Command("register", "Register a new user.")
	serve    = app.Command("serve", "Start the server.")
)

func init() {
	dbPath = os.Getenv("QUIMBY_DB")
	static = os.Getenv("QUIMBY_STATIC")

	// if len(static) == 0 {
	// 	log.Fatal("you must set the GADGETS_STATIC env var")
	// }
}

func main() {
	port := os.Getenv("QUIMBY_PORT")
	if port == "" {
		log.Fatal("you must specify a port with QUIMBY_PORT")
	}
	pth := os.Getenv("QUIMBY_DB")
	if pth == "" {
		log.Fatal("you must specify a db location with QUIMBY_DB")
	}
	db, err := models.GetDB(pth)
	if err != nil {
		log.Fatal(err)
	}
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// Register user
	case register.FullCommand():
		registerUser(db)
	case serve.FullCommand():
		auth.DB = db
		start(db, port, "/")
	}
	defer db.Close()

}

func start(db *bolt.DB, port, root string) {
	auth.DB = db

	r := mux.NewRouter()

	r.HandleFunc("/api/login", auth.Login).Methods("POST")
	r.HandleFunc("/api/logout", auth.Logout).Methods("POST")

	r.HandleFunc("/api/ping", Ping).Methods("GET")
	r.HandleFunc("/api/users/current", GetUser).Methods("GET")
	r.HandleFunc("/api/gadgets", GetGadgets).Methods("GET")
	r.HandleFunc("/api/gadgets", AddGadget).Methods("POST")
	r.HandleFunc("/api/gadgets/{name}", GetGadget).Methods("GET")
	r.HandleFunc("/api/gadgets/{name}", SendCommand).Methods("POST")
	r.HandleFunc("/api/gadgets/{name}", DeleteGadget).Methods("DELETE")
	r.HandleFunc("/api/gadgets/{name}/updates", Connect).Methods("GET")
	r.HandleFunc("/api/gadgets/{name}/status", GetStatus).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(rice.MustFindBox("www/app").HTTPBox()))
	//r.PathPrefix("/").Handler(http.FileServer(http.Dir(static)))

	http.Handle(root, r)
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("listening on %s", addr)
	err := http.ListenAndServe(addr, r)
	log.Println(err)
}

func Ping(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.Ping, auth.Read)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.GetUser, auth.Read)
}

func GetGadgets(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.GetGadgets, auth.Read)
}

func GetGadget(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.GetGadget, auth.Read)
}

func AddGadget(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.AddGadget, auth.Write)
}

func SendCommand(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.SendCommand, auth.Write)
}

func DeleteGadget(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.DeleteGadget, auth.Write)
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.GetStatus, auth.Read)
}

func Connect(w http.ResponseWriter, r *http.Request) {
	auth.CheckAuth(w, r, controllers.Connect, auth.Write)
}

func registerUser(db *bolt.DB) {
	u := models.User{
		DB: db,
	}
	fmt.Print("username: ")
	fmt.Scanf("%s", &u.Username)
	fmt.Print("can write? (y/N): ")
	var perm string
	fmt.Scanf("%s", &perm)
	if perm == "y" || perm == "Y" {
		u.Permission = "write"
	}
	fmt.Printf("password: ")
	p1 := string(gopass.GetPasswd())
	fmt.Printf("again: ")
	p2 := string(gopass.GetPasswd())
	if p1 != p2 {
		log.Fatal("passwords don't match")
	}
	u.Password = p1
	log.Println(u.Save())
}