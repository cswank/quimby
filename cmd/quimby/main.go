package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/GeertJohan/go.rice"
	"github.com/boltdb/bolt"
	"github.com/cswank/quimby"
	"github.com/cswank/quimby/cmd/quimby/handlers"
	"github.com/cswank/quimby/cmd/quimby/utils"
	"github.com/cswank/rex"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	version = "0.3.0"
)

var (
	users        = kingpin.Command("users", "User management")
	userAdd      = users.Command("add", "Add a new user.")
	userList     = users.Command("list", "List users.")
	userEdit     = users.Command("edit", "Update a user.")
	cert         = kingpin.Command("cert", "Make an tls cert.")
	domain       = cert.Flag("domain", "The domain for the tls cert.").Required().Short('d').String()
	pth          = cert.Flag("path", "The directory where the cert files will be written").Required().Short('p').String()
	serve        = kingpin.Command("serve", "Start the server.")
	command      = kingpin.Command("command", "Send a command.")
	method       = kingpin.Command("method", "Send a method.")
	gadgets      = kingpin.Command("gadgets", "Commands for managing gadgets")
	gadgetAdd    = gadgets.Command("add", "Add a gadget.")
	gadgetList   = gadgets.Command("list", "List the gadgets.")
	gadgetEdit   = gadgets.Command("edit", "List the gadgets.")
	gadgetDelete = gadgets.Command("delete", "Delete a gadget.")
	token        = kingpin.Command("token", "Generate a jwt token")
	bootstrap    = kingpin.Command("bootstrap", "Set up a bunch of stuff")

	keyPath  = os.Getenv("QUIMBY_TLS_KEY")
	certPath = os.Getenv("QUIMBY_TLS_CERT")
	iface    = os.Getenv("QUIMBY_INTERFACE")
)

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("Craig Swank")
	switch kingpin.Parse() {
	case "cert":
		utils.GenerateCert(*domain, *pth)
	case "users add":
		addDB(utils.AddUser)
	case "users list":
		addDB(utils.ListUsers)
	case "users edit":
		addDB(utils.EditUser)
	case "gadgets add":
		addDB(utils.AddGadget)
	case "gadgets list":
		addDB(utils.ListGadgets)
	case "gadgets edit":
		addDB(utils.EditGadget)
	case "gadgets delete":
		addDB(utils.DeleteGadget)
	case "command":
		addDB(utils.SendCommand)
	case "token":
		utils.GetToken()
	case "bootstrap":
		utils.Bootstrap()
	case "serve":
		addDB(startServer)
	}
}

type dbNeeder func(*bolt.DB)

func addDB(f dbNeeder) {
	pth := os.Getenv("QUIMBY_DB")
	if pth == "" {
		log.Fatal("you must specify a db location with QUIMBY_DB")
	}
	db, err := quimby.GetDB(pth)
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
	clients := quimby.NewClientHolder()
	start(db, port, internalPort, "/", "/api", lg, clients)
}

func start(db *bolt.DB, port, internalPort, root string, iRoot string, lg quimby.Logger, clients *quimby.ClientHolder) {
	quimby.Clients = clients
	quimby.DB = db
	quimby.LG = lg
	handlers.DB = db
	handlers.LG = lg

	go startInternal(iRoot, lg, internalPort)

	r := rex.New("main")
	r.Post("/api/login", handlers.Login)
	r.Post("/api/logout", handlers.Logout)
	r.Get("/api/ping", ping)
	r.Get("/api/users/current", getUser)
	r.Get("/api/gadgets", getGadgets)
	r.Post("/api/gadgets", addGadget)
	r.Get("/api/gadgets/{id}", getGadget)
	r.Post("/api/gadgets/{id}", sendCommand)
	r.Delete("/api/gadgets/{id}", deleteGadget)
	r.Post("/api/gadgets/{id}/method", sendMethod)
	r.Get("/api/gadgets/{id}/websocket", connect)
	r.Get("/api/gadgets/{id}/values", getValues)
	r.Get("/api/gadgets/{id}/status", getStatus)
	r.Post("/api/gadgets/{id}/notes", addNote)
	r.Get("/api/gadgets/{id}/notes", getNotes)
	r.Get("/api/gadgets/{id}/locations/{location}/devices/{device}/status", getDevice)
	r.Post("/api/gadgets/{id}/locations/{location}/devices/{device}/status", updateDevice)
	r.Get("/api/gadgets/{id}/sources/{name}", getDataPoints)
	r.Get("/api/gadgets/{id}/sources/{name}/csv", getDataPointsCSV)
	r.Get("/admin/clients", getClients)

	r.ServeFiles(http.FileServer(rice.MustFindBox("www/dist").HTTPBox()))

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
func startInternal(iRoot string, lg quimby.Logger, port string) {
	r := rex.New("internal")
	r.Post("/internal/updates", relay)
	r.Post("/internal/gadgets/{id}/sources/{name}", addDataPoint)
	http.Handle(iRoot, r)
	a := fmt.Sprintf(":%s", port)
	lg.Printf("listening on %s", a)
	err := http.ListenAndServe(a, r)
	if err != nil {
		log.Fatal(err)
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.Ping, handlers.Read, "main")
}

func getUser(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetUser, handlers.Read, "main")
}

func getGadgets(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetGadgets, handlers.Read, "main")
}

func gadgetOptions(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GadgetOptions, handlers.Read, "main")
}

func getGadget(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetGadget, handlers.Read, "main")
}

func addGadget(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.AddGadget, handlers.Write, "main")
}

func sendCommand(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.SendCommand, handlers.Write, "main")
}

func sendMethod(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.SendMethod, handlers.Write, "main")
}

func deleteGadget(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.DeleteGadget, handlers.Write, "main")
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetStatus, handlers.Read, "main")
}

func addNote(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.AddNote, handlers.Write, "main")
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetNotes, handlers.Read, "main")
}

func addDataPoint(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.AddDataPoint, handlers.Write, "internal")
}

func getDataPoints(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetDataPoints, handlers.Read, "main")
}

func getDataPointsCSV(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetDataPointsCSV, handlers.Read, "main")
}

func getValues(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetValues, handlers.Read, "main")
}

func connect(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.Connect, handlers.Read, "main")
}

func relay(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.RelayMessage, handlers.Write, "internal")
}
func updateDevice(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.UpdateDevice, handlers.Write, "main")
}

func getDevice(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetDevice, handlers.Write, "main")
}

func getClients(w http.ResponseWriter, r *http.Request) {
	handlers.Handle(w, r, handlers.GetClients, handlers.Admin, "main")
}
