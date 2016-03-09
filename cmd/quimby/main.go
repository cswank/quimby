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
	"github.com/justinas/alice"
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

	go startInternal(iRoot, db, lg, internalPort)

	r := rex.New("main")
	r.Post("/api/login", http.HandlerFunc(handlers.Login))
	r.Post("/api/logout", http.HandlerFunc(handlers.Logout))
	r.Get("/api/ping", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.Ping)))
	r.Get("/api/users/current", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetUser)))
	r.Get("/api/gadgets", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetGadgets)))
	r.Post("/api/gadgets", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.AddGadget)))
	r.Get("/api/gadgets/{id}", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetGadget)))
	r.Post("/api/gadgets/{id}", alice.New(handlers.Perm(handlers.Write)).Then(http.HandlerFunc(handlers.SendCommand)))
	r.Delete("/api/gadgets/{id}", alice.New(handlers.Perm(handlers.Write)).Then(http.HandlerFunc(handlers.DeleteGadget)))
	r.Post("/api/gadgets/{id}/method", alice.New(handlers.Perm(handlers.Write)).Then(http.HandlerFunc(handlers.SendMethod)))
	r.Get("/api/gadgets/{id}/websocket", alice.New(handlers.Perm(handlers.Write)).Then(http.HandlerFunc(handlers.Connect)))
	r.Get("/api/gadgets/{id}/values", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetValues)))
	r.Get("/api/gadgets/{id}/status", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetStatus)))
	r.Post("/api/gadgets/{id}/notes", alice.New(handlers.Perm(handlers.Write)).Then(http.HandlerFunc(handlers.AddNote)))
	r.Get("/api/gadgets/{id}/notes", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetNotes)))
	r.Get("/api/gadgets/{id}/locations/{location}/devices/{device}/status", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetDevice)))
	r.Post("/api/gadgets/{id}/locations/{location}/devices/{device}/status", alice.New(handlers.Perm(handlers.Write)).Then(http.HandlerFunc(handlers.UpdateDevice)))
	r.Get("/api/gadgets/{id}/sources/{name}", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetDataPoints)))
	r.Get("/api/gadgets/{id}/sources/{name}/csv", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetDataPointsCSV)))
	r.Get("/beer/{name}", alice.New(handlers.Perm(handlers.Read)).Then(http.HandlerFunc(handlers.GetRecipe)))
	r.Get("/admin/clients", alice.New(handlers.Perm(handlers.Admin)).Then(http.HandlerFunc(handlers.GetClients)))

	r.ServeFiles(http.FileServer(rice.MustFindBox("www/dist").HTTPBox()))

	chain := alice.New(handlers.Auth(db, lg, r, "main"), handlers.FetchGadget()).Then(r)

	http.Handle(root, chain)

	addr := fmt.Sprintf("%s:%s", iface, port)
	lg.Printf("listening on %s\n", addr)
	if keyPath == "" {
		lg.Println(http.ListenAndServe(addr, chain))
	} else {
		lg.Println(http.ListenAndServeTLS(fmt.Sprintf("%s:443", iface), certPath, keyPath, nil))
	}
}

//This is the endpoint that the gadgets report to. It is
//served on a separate port so it doesn't have to be exposed
//publicly if the main port is exposed.
func startInternal(iRoot string, db *bolt.DB, lg quimby.Logger, port string) {
	r := rex.New("internal")
	r.Post("/internal/updates", alice.New(handlers.Perm(handlers.Write)).Then(http.HandlerFunc(handlers.RelayMessage)))
	r.Post("/internal/gadgets/{id}/sources/{name}", alice.New(handlers.Perm(handlers.Write)).Then(http.HandlerFunc(handlers.AddDataPoint)))

	chain := alice.New(handlers.Auth(db, lg, r, "internal"), handlers.FetchGadget()).Then(r)

	http.Handle(iRoot, chain)
	a := fmt.Sprintf(":%s", port)
	lg.Printf("listening on %s", a)
	err := http.ListenAndServe(a, chain)
	if err != nil {
		log.Fatal(err)
	}
}
