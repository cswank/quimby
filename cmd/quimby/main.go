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
	version = "0.5.0"
)

var (
	users        = kingpin.Command("users", "User management")
	userAdd      = users.Command("add", "Add a new user.")
	userName     = users.Flag("username", "Username for a new user").String()
	userPW       = users.Flag("password", "Password for a new user").String()
	userPerm     = users.Flag("permission", "Permission (read, write, or admin").String()
	userList     = users.Command("list", "List users.")
	userEdit     = users.Command("edit", "Update a user.")
	cert         = kingpin.Command("cert", "Make an tls cert.")
	domain       = cert.Flag("domain", "The domain for the tls cert.").Required().Short('d').String()
	pth          = cert.Flag("path", "The directory where the cert files will be written").Required().Short('p').String()
	serve        = kingpin.Command("serve", "Start the server.")
	setup        = kingpin.Command("setup", "Set up the the server (keys and init scripts and what not.")
	net          = setup.Flag("net", "network interface").Short('n').Default("eth0").String()
	setupDomain  = setup.Flag("domain", "network interface").Required().Short('d').String()
	command      = kingpin.Command("command", "Send a command.")
	method       = kingpin.Command("method", "Send a method.")
	gadgets      = kingpin.Command("gadgets", "Commands for managing gadgets")
	gadgetAdd    = gadgets.Command("add", "Add a gadget.")
	gadgetName   = gadgets.Flag("name", "Name of the gadget.").String()
	gadgetHost   = gadgets.Flag("host", "ip address of gadget (id http://<ipaddr>:6111)").String()
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
		doUser(utils.AddUser)
	case "users list":
		addDB(utils.ListUsers)
	case "users edit":
		addDB(utils.EditUser)
	case "gadgets add":
		doGadget(utils.AddGadget)
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
	case "setup":
		utils.SetupServer(*setupDomain, *net)
	}
}

type dbNeeder func(*bolt.DB)
type userNeeder func(*quimby.User)
type gadgetNeeder func(*quimby.Gadget)

func getDB() *bolt.DB {
	pth := os.Getenv("QUIMBY_DB")
	if pth == "" {
		log.Fatal("you must specify a db location with QUIMBY_DB")
	}
	db, err := quimby.GetDB(pth)
	if err != nil {
		log.Fatalf("could not open db at %s - %v", pth, err)
	}
	return db
}

func doUser(f userNeeder) {
	db := getDB()
	u := &quimby.User{
		DB:         db,
		Password:   *userPW,
		Username:   *userName,
		Permission: *userPerm,
	}
	f(u)
	defer db.Close()
}

func doGadget(f gadgetNeeder) {
	db := getDB()
	g := &quimby.Gadget{
		DB:   db,
		Name: *gadgetName,
		Host: *gadgetHost,
	}
	f(g)
	db.Close()
}

func addDB(f dbNeeder) {
	db := getDB()
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

func getMiddleware(perm handlers.ACL, f http.HandlerFunc) http.Handler {
	return alice.New(handlers.Perm(perm)).Then(http.HandlerFunc(f))
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
	r.Get("/api/ping", getMiddleware(handlers.Read, handlers.Ping))
	r.Get("/api/currentuser", getMiddleware(handlers.Read, handlers.GetCurrentUser))
	r.Get("/api/users", getMiddleware(handlers.Admin, handlers.GetUsers))
	r.Get("/api/users/{username}", getMiddleware(handlers.Admin, handlers.GetUser))
	r.Get("/api/gadgets", getMiddleware(handlers.Read, handlers.GetGadgets))
	r.Post("/api/gadgets", getMiddleware(handlers.Read, handlers.AddGadget))
	r.Get("/api/gadgets/{id}", getMiddleware(handlers.Read, handlers.GetGadget))
	r.Post("/api/gadgets/{id}", getMiddleware(handlers.Write, handlers.UpdateGadget))
	r.Delete("/api/gadgets/{id}", getMiddleware(handlers.Write, handlers.DeleteGadget))
	r.Post("/api/gadgets/{id}/command", getMiddleware(handlers.Write, handlers.SendCommand))
	r.Post("/api/gadgets/{id}/method", getMiddleware(handlers.Write, handlers.SendMethod))
	r.Get("/api/gadgets/{id}/websocket", getMiddleware(handlers.Write, handlers.Connect))
	r.Get("/api/gadgets/{id}/values", getMiddleware(handlers.Read, handlers.GetUpdates))
	r.Get("/api/gadgets/{id}/status", getMiddleware(handlers.Read, handlers.GetStatus))
	r.Post("/api/gadgets/{id}/notes", getMiddleware(handlers.Write, handlers.AddNote))
	r.Get("/api/gadgets/{id}/notes", getMiddleware(handlers.Read, handlers.GetNotes))
	r.Get("/api/gadgets/{id}/locations/{location}/devices/{device}/status", getMiddleware(handlers.Read, handlers.GetDevice))
	r.Post("/api/gadgets/{id}/locations/{location}/devices/{device}/status", getMiddleware(handlers.Write, handlers.UpdateDevice))
	r.Get("/api/gadgets/{id}/sources/{name}", getMiddleware(handlers.Read, handlers.GetDataPoints))
	r.Get("/api/gadgets/{id}/sources/{name}/csv", getMiddleware(handlers.Read, handlers.GetDataPointsCSV))
	r.Get("/api/beer/{name}", getMiddleware(handlers.Read, handlers.GetRecipe))
	r.Get("/api/admin/clients", getMiddleware(handlers.Admin, handlers.GetClients))

	r.ServeFiles(http.FileServer(rice.MustFindBox("www/dist").HTTPBox()))

	chain := alice.New(handlers.Auth(db, lg, "main"), handlers.FetchGadget(), handlers.Error(lg)).Then(r)

	http.Handle(root, chain)

	addr := fmt.Sprintf("%s:%s", iface, port)
	lg.Printf("listening on %s\n", addr)
	if keyPath == "" {
		lg.Println(http.ListenAndServe(addr, chain))
	} else {
		lg.Println(http.ListenAndServeTLS(fmt.Sprintf("%s:443", iface), certPath, keyPath, chain))
	}
}

//This is the endpoint that the gadgets report to. It is
//served on a separate port so it doesn't have to be exposed
//publicly if the main port is exposed.
func startInternal(iRoot string, db *bolt.DB, lg quimby.Logger, port string) {
	r := rex.New("internal")
	r.Post("/internal/updates", getMiddleware(handlers.Write, handlers.RelayMessage))
	r.Post("/internal/gadgets/{id}/sources/{name}", getMiddleware(handlers.Write, handlers.AddDataPoint))

	chain := alice.New(handlers.Auth(db, lg, "internal"), handlers.FetchGadget()).Then(r)

	http.Handle(iRoot, chain)
	a := fmt.Sprintf(":%s", port)
	lg.Printf("listening on %s", a)
	err := http.ListenAndServe(a, chain)
	if err != nil {
		log.Fatal(err)
	}
}
