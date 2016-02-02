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
	clients := models.NewClientHolder()
	start(db, port, internalPort, "/", "/api", lg, clients)
}

func start(db *bolt.DB, port, internalPort, root string, iRoot string, lg models.Logger, clients *models.ClientHolder) {
	models.Clients = clients
	models.DB = db
	models.LG = lg
	controllers.DB = db
	controllers.LG = lg

	go startInternal(iRoot, lg, internalPort)

	r := rex.New("main")
	r.Post("/api/login", controllers.Login)
	r.Post("/api/logout", controllers.Logout)
	r.Get("/api/ping", Ping)
	r.Get("/api/users/current", GetUser)
	r.Get("/api/gadgets", GetGadgets)
	r.Post("/api/gadgets", AddGadget)
	r.Get("/api/gadgets/{id}", GetGadget)
	r.Post("/api/gadgets/{id}", SendCommand)
	r.Delete("/api/gadgets/{id}", DeleteGadget)
	r.Post("/api/gadgets/{id}/method", SendMethod)
	r.Get("/api/gadgets/{id}/websocket", Connect)
	r.Get("/api/gadgets/{id}/values", GetValues)
	r.Get("/api/gadgets/{id}/status", GetStatus)
	r.Post("/api/gadgets/{id}/notes", AddNote)
	r.Get("/api/gadgets/{id}/notes", GetNotes)
	r.Get("/api/gadgets/{id}/locations/{location}/devices/{device}/status", GetDevice)
	r.Post("/api/gadgets/{id}/locations/{location}/devices/{device}/status", UpdateDevice)
	r.Get("/api/gadgets/{id}/sources/{name}", GetDataPoints)
	r.Get("/api/gadgets/{id}/sources/{name}/csv", GetDataPointsCSV)
	r.Get("/admin/clients", GetClients)

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
func startInternal(iRoot string, lg models.Logger, port string) {
	r := rex.New("internal")
	r.Post("/internal/updates", Relay)
	r.Post("/internal/gadgets/{id}/sources/{name}", AddDataPoint)
	http.Handle(iRoot, r)
	a := fmt.Sprintf(":%s", port)
	lg.Printf("listening on %s", a)
	err := http.ListenAndServe(a, r)
	if err != nil {
		log.Fatal(err)
	}
}

func Ping(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.Ping, controllers.Read, "main")
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetUser, controllers.Read, "main")
}

func GetGadgets(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetGadgets, controllers.Read, "main")
}

func GadgetOptions(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GadgetOptions, controllers.Read, "main")
}

func GetGadget(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetGadget, controllers.Read, "main")
}

func AddGadget(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.AddGadget, controllers.Write, "main")
}

func SendCommand(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.SendCommand, controllers.Write, "main")
}

func SendMethod(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.SendMethod, controllers.Write, "main")
}

func DeleteGadget(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.DeleteGadget, controllers.Write, "main")
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetStatus, controllers.Read, "main")
}

func AddNote(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.AddNote, controllers.Write, "main")
}

func GetNotes(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetNotes, controllers.Read, "main")
}

func AddDataPoint(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.AddDataPoint, controllers.Write, "internal")
}

func GetDataPoints(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetDataPoints, controllers.Read, "main")
}

func GetDataPointsCSV(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetDataPointsCSV, controllers.Read, "main")
}

func GetValues(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetValues, controllers.Read, "main")
}

func Connect(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.Connect, controllers.Read, "main")
}

func Relay(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.RelayMessage, controllers.Write, "internal")
}
func UpdateDevice(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.UpdateDevice, controllers.Write, "main")
}

func GetDevice(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetDevice, controllers.Write, "main")
}

func GetClients(w http.ResponseWriter, r *http.Request) {
	controllers.Handle(w, r, controllers.GetClients, controllers.Admin, "main")
}
