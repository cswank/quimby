package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/GeertJohan/go.rice"
	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/admin"
	"github.com/cswank/quimby/controllers"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	dbPath       string
	app          = kingpin.New("quimby", "An interface to gogadets")
	users        = app.Command("users", "User management")
	userAdd      = users.Command("add", "Add a new user.")
	userList     = users.Command("list", "List users.")
	serve        = app.Command("serve", "Start the server.")
	gadgets      = app.Command("gadgets", "Commands for managing gadgets")
	gadgetAdd    = gadgets.Command("add", "Add a gadget.")
	gadgetList   = gadgets.Command("list", "List the gadgets.")
	gadgetEdit   = gadgets.Command("edit", "List the gadgets.")
	gadgetDelete = gadgets.Command("delete", "Delete a gadget.")

	keyPath  = os.Getenv("QUIMBY_TLS_KEY")
	certPath = os.Getenv("QUIMBY_TLS_CERT")
	iface    = os.Getenv("QUIMBY_INTERFACE")
)

func init() {
	dbPath = os.Getenv("QUIMBY_DB")
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
	lg := log.New(os.Stdout, "quimby ", log.Ltime)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case userAdd.FullCommand():
		admin.AddUser(db)
	case userList.FullCommand():
		admin.ListUsers(db)
	case gadgetAdd.FullCommand():
		admin.AddGadget(db)
	case gadgetList.FullCommand():
		admin.ListGadgets(db)
	case gadgetEdit.FullCommand():
		admin.EditGadget(db)
	case gadgetDelete.FullCommand():
		admin.DeleteGadget(db)
	case serve.FullCommand():
		start(db, port, "/", "/api", lg)
	}
	defer db.Close()
}

func start(db *bolt.DB, port, root string, iRoot string, lg controllers.Logger) {
	controllers.DB = db
	controllers.LG = lg

	go startInternal(iRoot, lg)

	r := mux.NewRouter()

	r.HandleFunc("/api/login", controllers.Login).Methods("POST")
	r.HandleFunc("/api/logout", controllers.Logout).Methods("POST")
	r.HandleFunc("/api/ping", Ping).Methods("GET")
	r.HandleFunc("/api/users/current", GetUser).Methods("GET")
	r.HandleFunc("/api/gadgets", GetGadgets).Methods("GET")
	r.HandleFunc("/api/gadgets", AddGadget).Methods("POST")
	r.HandleFunc("/api/gadgets/{id}", GetGadget).Methods("GET")
	r.HandleFunc("/api/gadgets/{id}", SendCommand).Methods("POST")
	r.HandleFunc("/api/gadgets/{id}", DeleteGadget).Methods("DELETE")
	r.HandleFunc("/api/gadgets/{id}/websocket", GetTicket).Methods("GET")
	r.HandleFunc("/api/websocket/{ticket}", Connect).Methods("GET")
	r.HandleFunc("/api/gadgets/{id}/values", GetValues).Methods("GET")
	r.HandleFunc("/api/gadgets/{id}/status", GetStatus).Methods("GET")
	r.HandleFunc("/api/gadgets/{id}/locations/{location}/devices/{device}/status", GetDevice).Methods("GET")
	r.HandleFunc("/api/gadgets/{id}/locations/{location}/devices/{device}/status", UpdateDevice).Methods("POST")

	r.PathPrefix("/").Handler(http.FileServer(rice.MustFindBox("www/dist").HTTPBox()))

	http.Handle(root, r)
	addr := fmt.Sprintf("%s:%s", iface, port)
	lg.Printf("listening on %s\n", addr)
	if keyPath == "" {
		lg.Println(http.ListenAndServe(addr, r))
	} else {
		lg.Println(http.ListenAndServeTLS(fmt.Sprintf("%s:443", iface), certPath, keyPath, nil))
	}
}

//This is the endpoint that the gadgets report to. It is
//served on a separate port so it doesn't have to be exposed
//publicly if the main port is exposed.
func startInternal(iRoot string, lg controllers.Logger) {
	r := mux.NewRouter()
	r.HandleFunc("/internal/updates", Relay).Methods("POST")
	http.Handle(iRoot, r)
	a := fmt.Sprintf(":%s", os.Getenv("QUIMBY_INTERNAL_PORT"))
	lg.Printf("listening on %s", a)
	err := http.ListenAndServe(a, r)
	if err != nil {
		log.Fatal(err)
	}
}

func Ping(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.Ping, controllers.Read)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetUser, controllers.Read)
}

func GetGadgets(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetGadgets, controllers.Read)
}

func GetGadget(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetGadget, controllers.Read)
}

func AddGadget(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.AddGadget, controllers.Write)
}

func SendCommand(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.SendCommand, controllers.Write)
}

func DeleteGadget(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.DeleteGadget, controllers.Write)
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetStatus, controllers.Read)
}

func GetValues(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetValues, controllers.Read)
}

func GetTicket(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetTicket, controllers.Read)
}

func Connect(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.Connect, controllers.Read)
}

func Relay(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.RelayMessage, controllers.Write)
}

func UpdateDevice(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.UpdateDevice, controllers.Write)
}

func GetDevice(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetDevice, controllers.Write)
}
