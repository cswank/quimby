package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/GeertJohan/go.rice"
	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/controllers"
	"github.com/cswank/quimby/models"
	"github.com/cswank/quimby/utils"
	"github.com/go-zoo/bone"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app          = kingpin.New("quimby", "An interface to gogadets")
	users        = app.Command("users", "User management")
	userAdd      = users.Command("add", "Add a new user.")
	userList     = users.Command("list", "List users.")
	userEdit     = users.Command("edit", "Update a user.")
	cert         = app.Command("cert", "Make an ssl cert.")
	domain       = cert.Flag("domain", "The domain for the tls cert.").Required().Short('d').String()
	pth          = cert.Flag("path", "The directory where the cert files will be written").Required().Short('p').String()
	serve        = app.Command("serve", "Start the server.")
	command      = app.Command("command", "Start the server.")
	gadgets      = app.Command("gadgets", "Commands for managing gadgets")
	gadgetAdd    = gadgets.Command("add", "Add a gadget.")
	gadgetList   = gadgets.Command("list", "List the gadgets.")
	gadgetEdit   = gadgets.Command("edit", "List the gadgets.")
	gadgetDelete = gadgets.Command("delete", "Delete a gadget.")

	keyPath  = os.Getenv("QUIMBY_TLS_KEY")
	certPath = os.Getenv("QUIMBY_TLS_CERT")
	iface    = os.Getenv("QUIMBY_INTERFACE")
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case cert.FullCommand():
		utils.GenerateCert(*domain, *pth)
	case userAdd.FullCommand():
		addDB(utils.AddUser)
	case userList.FullCommand():
		addDB(utils.ListUsers)
	case userEdit.FullCommand():
		addDB(utils.EditUser)
	case gadgetAdd.FullCommand():
		addDB(utils.AddGadget)
	case gadgetList.FullCommand():
		addDB(utils.ListGadgets)
	case gadgetEdit.FullCommand():
		addDB(utils.EditGadget)
	case gadgetDelete.FullCommand():
		addDB(utils.DeleteGadget)
	case command.FullCommand():
		addDB(utils.SendCommand)
	case serve.FullCommand():
		addDB(startServer)
	}
}

type dbNeeder func(*bolt.DB)

func addDB(f dbNeeder) {
	pth := os.Getenv("QUIMBY_DB")
	if pth == "" {
		log.Fatal("you must specify a db location with QUIMBY_DB")
	}
	db, err := models.GetDB(pth)
	if err != nil {
		log.Fatalf("could not open db at %s - %v", pth, err)
	}
	f(db)
	defer db.Close()
}

func startServer(db *bolt.DB) {
	port := os.Getenv("QUIMBY_PORT")
	if port == "" {
		log.Fatal("you must specify a port with QUIMBY_PORT")
	}

	internalPort := os.Getenv("QUIMBY_INTERNAL_PORT")
	if port == "" {
		log.Fatal("you must specify a port with QUIMBY_INTERNAL_PORT")
	}
	lg := log.New(os.Stdout, "quimby ", log.Ltime)
	clients := controllers.NewClientHolder()
	start(db, port, internalPort, "/", "/api", lg, clients)
}

func start(db *bolt.DB, port, internalPort, root string, iRoot string, lg controllers.Logger, clients *controllers.ClientHolder) {
	controllers.Clients = clients
	controllers.DB = db
	controllers.LG = lg

	go startInternal(iRoot, lg, internalPort)

	mux := bone.New()

	mux.Post("/api/login", http.HandlerFunc(controllers.Login))
	mux.Post("/api/logout", http.HandlerFunc(controllers.Logout))
	mux.Get("/api/ping", http.HandlerFunc(Ping))
	mux.Get("/api/users/current", http.HandlerFunc(GetUser))
	mux.Get("/api/gadgets", http.HandlerFunc(GetGadgets))
	//mux.Options("/api/:rest:.*}", http.HandlerFunc(GadgetsOptions))
	mux.Post("/api/gadgets", http.HandlerFunc(AddGadget))
	mux.Get("/api/gadgets/:id", http.HandlerFunc(GetGadget))
	mux.Post("/api/gadgets/:id", http.HandlerFunc(SendCommand))
	mux.Delete("/api/gadgets/:id", http.HandlerFunc(DeleteGadget))
	mux.Get("/api/gadgets/:id/websocket", http.HandlerFunc(Connect))
	mux.Get("/api/gadgets/:id/values", http.HandlerFunc(GetValues))
	mux.Get("/api/gadgets/:id/status", http.HandlerFunc(GetStatus))
	mux.Get("/api/gadgets/:id/locations/:location/devices/:device/status", http.HandlerFunc(GetDevice))
	mux.Post("/api/gadgets/:id/locations/:location/devices/:device/status", http.HandlerFunc(UpdateDevice))
	mux.Get("/admin/clients", http.HandlerFunc(GetClients))

	mux.Get("/", http.FileServer(rice.MustFindBox("www/dist").HTTPBox()))

	http.Handle(root, mux)
	addr := fmt.Sprintf("%s:%s", iface, port)
	lg.Printf("listening on %s\n", addr)
	if keyPath == "" {
		lg.Println(http.ListenAndServe(addr, mux))
	} else {
		lg.Println(http.ListenAndServeTLS(fmt.Sprintf("%s:443", iface), certPath, keyPath, nil))
	}
}

//This is the endpoint that the gadgets report to. It is
//served on a separate port so it doesn't have to be exposed
//publicly if the main port is exposed.
func startInternal(iRoot string, lg controllers.Logger, port string) {
	mux := bone.New()
	mux.Post("/internal/updates", http.HandlerFunc(Relay))
	http.Handle(iRoot, mux)
	a := fmt.Sprintf(":%s", port)
	lg.Printf("listening on %s", a)
	err := http.ListenAndServe(a, mux)
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

func GadgetsOptions(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.Options, controllers.Read)
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

func GetClients(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetClients, controllers.Admin)
}
