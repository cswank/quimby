package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/GeertJohan/go.rice"
	"github.com/boltdb/bolt"
	"github.com/cswank/quimby"
	"github.com/cswank/quimby/cmd/quimby/handlers"
	"github.com/cswank/quimby/cmd/quimby/handlers/webapp"
	"github.com/cswank/quimby/cmd/quimby/utils"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	version = "1.1.0"
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

	box *rice.Box
)

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("Craig Swank")
	switch kingpin.Parse() {
	case "cert":
		utils.GenerateCert(*domain, *pth)
	case "users add":
		doUser(utils.AddUser)
	case "users list":
		utils.ListUsers()
	case "users edit":
		utils.EditUser()
	case "gadgets add":
		doGadget(utils.AddGadget)
	case "gadgets list":
		utils.ListGadgets()
	case "gadgets edit":
		utils.EditGadget()
	case "gadgets delete":
		utils.DeleteGadget()
	case "command":
		utils.SendCommand()
	case "token":
		utils.GetToken()
	case "bootstrap":
		utils.Bootstrap()
	case "serve":
		box = rice.MustFindBox("static")
		webapp.Init(box)
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
	u := quimby.NewUser(
		*userName,
		quimby.UserPassword(*userPW),
		quimby.UserPermission(*userPerm),
	)
	f(u)
}

func doGadget(f gadgetNeeder) {
	g := &quimby.Gadget{
		Name: *gadgetName,
		Host: *gadgetHost,
	}
	f(g)
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

	domain := os.Getenv("QUIMBY_DOMAIN")
	if domain == "" {
		log.Fatal("you must specify a domain with QUIMBY_DOMAIN")
	}

	internalPort := os.Getenv("QUIMBY_INTERNAL_PORT")
	if port == "" {
		log.Fatal("you must specify a port with QUIMBY_INTERNAL_PORT")
	}

	var lg *log.Logger
	if os.Getenv("QUIMBY_NULLLOG") != "" {
		lg = log.New(ioutil.Discard, "quimby ", log.Ltime)
	} else {
		lg = log.New(os.Stdout, "quimby ", log.Ltime)
	}
	clients := quimby.NewClientHolder()
	tfa := quimby.NewTFA(domain)
	start(db, port, internalPort, "/", "/api", lg, clients, tfa)
}

func getMiddleware(perm handlers.ACL, f handlers.HandlerFunc) http.Handler {
	return alice.New(handlers.Auth(), handlers.FetchGadget(), handlers.Perm(perm)).Then(http.HandlerFunc(handlers.Error(f)))
}

func start(db *bolt.DB, port, internalPort, root string, iRoot string, lg quimby.Logger, clients *quimby.ClientHolder, tfa quimby.TFAer) {
	quimby.Clients = clients
	quimby.SetDB(db)
	quimby.LG = lg
	handlers.LG = lg
	handlers.TFA = tfa

	go startInternal(iRoot, db, lg, internalPort)
	go startHomeKit(db, lg)

	r := mux.NewRouter()
	r.Handle("/home", getMiddleware(handlers.Read, webapp.IndexPage)).Methods("GET")
	r.Handle("/gadgets/{id}", getMiddleware(handlers.Read, webapp.GadgetPage)).Methods("GET")
	r.Handle("/gadgets/{id}/method.html", getMiddleware(handlers.Read, webapp.EditMethodPage)).Methods("GET")
	r.Handle("/gadgets/{id}/chart.html", getMiddleware(handlers.Read, webapp.ChartPage)).Methods("GET")
	r.Handle("/gadgets/{id}/chart-setup.html", getMiddleware(handlers.Read, webapp.ChartSetupPage)).Methods("GET")
	r.Handle("/gadgets/{id}/chart-setup/{name}", getMiddleware(handlers.Read, webapp.ChartInputPage)).Methods("GET")
	r.Handle("/login.html", getMiddleware(handlers.Anyone, webapp.LoginPage)).Methods("GET")
	r.Handle("/login.html", getMiddleware(handlers.Anyone, webapp.LoginForm)).Methods("POST")
	r.Handle("/logout.html", getMiddleware(handlers.Read, webapp.LogoutPage)).Methods("GET")
	r.Handle("/links.html", getMiddleware(handlers.Read, webapp.LinksPage)).Methods("GET")
	r.Handle("/admin.html", getMiddleware(handlers.Admin, webapp.AdminPage)).Methods("GET")
	r.Handle("/admin/confirmation", getMiddleware(handlers.Admin, webapp.DeleteConfirmPage)).Methods("GET")
	r.Handle("/admin/gadgets/{gadgetid}", getMiddleware(handlers.Admin, webapp.GadgetEditPage)).Methods("GET")
	r.Handle("/admin/gadgets/{gadgetid}", getMiddleware(handlers.Admin, webapp.GadgetForm)).Methods("POST")
	r.Handle("/admin/gadgets/{gadgetid}", getMiddleware(handlers.Admin, webapp.DeleteGadgetPage)).Methods("DELETE")
	r.Handle("/admin/users/{username}", getMiddleware(handlers.Admin, webapp.UserEditPage)).Methods("GET")
	r.Handle("/admin/users/{username}/password", getMiddleware(handlers.Admin, webapp.UserPasswordPage)).Methods("GET")
	r.Handle("/admin/users/{username}/password", getMiddleware(handlers.Admin, webapp.UserChangePasswordPage)).Methods("POST")
	r.Handle("/admin/users/{username}/tfa", getMiddleware(handlers.Admin, webapp.UserTFAPage)).Methods("POST")
	r.Handle("/admin/users/{username}", getMiddleware(handlers.Admin, webapp.DeleteUserPage)).Methods("DELETE")
	r.Handle("/admin/users/{username}", getMiddleware(handlers.Admin, webapp.UserForm)).Methods("POST")

	//api
	r.Handle("/api/login", http.HandlerFunc(handlers.Login)).Methods("POST")
	r.Handle("/api/logout", http.HandlerFunc(handlers.Logout)).Methods("POST")
	r.Handle("/api/ping", getMiddleware(handlers.Read, handlers.Ping)).Methods("GET")
	r.Handle("/api/currentuser", getMiddleware(handlers.Read, handlers.GetCurrentUser)).Methods("GET")
	r.Handle("/api/users", getMiddleware(handlers.Admin, handlers.GetUsers)).Methods("GET")
	r.Handle("/api/users", getMiddleware(handlers.Admin, handlers.AddUser)).Methods("POST")
	r.Handle("/api/users/{username}", getMiddleware(handlers.Admin, handlers.DeleteUser))
	r.Handle("/api/users/{username}/permission", getMiddleware(handlers.Admin, handlers.UpdateUserPermission)).Methods("POST")
	r.Handle("/api/users/{username}/password", getMiddleware(handlers.Admin, handlers.UpdateUserPassword)).Methods("POST")
	r.Handle("/api/users/{username}", getMiddleware(handlers.Admin, handlers.GetUser)).Methods("GET")
	r.Handle("/api/gadgets", getMiddleware(handlers.Read, handlers.GetGadgets)).Methods("GET")
	r.Handle("/api/gadgets", getMiddleware(handlers.Read, handlers.AddGadget)).Methods("POST")
	r.Handle("/api/gadgets/{id}", getMiddleware(handlers.Read, handlers.GetGadget)).Methods("GET")
	r.Handle("/api/gadgets/{id}", getMiddleware(handlers.Write, handlers.UpdateGadget)).Methods("POST")
	r.Handle("/api/gadgets/{id}", getMiddleware(handlers.Write, handlers.DeleteGadget)).Methods("DELETE")
	r.Handle("/api/gadgets/{id}/command", getMiddleware(handlers.Write, handlers.SendCommand)).Methods("POST")
	r.Handle("/api/gadgets/{id}/method", getMiddleware(handlers.Write, handlers.SendMethod)).Methods("POST")
	r.Handle("/api/gadgets/{id}/websocket", getMiddleware(handlers.Write, handlers.Connect)).Methods("GET")
	r.Handle("/api/gadgets/{id}/values", getMiddleware(handlers.Read, handlers.GetUpdates)).Methods("GET")
	r.Handle("/api/gadgets/{id}/status", getMiddleware(handlers.Read, handlers.GetStatus)).Methods("GET")
	r.Handle("/api/gadgets/{id}/notes", getMiddleware(handlers.Write, handlers.AddNote)).Methods("POST")
	r.Handle("/api/gadgets/{id}/notes", getMiddleware(handlers.Read, handlers.GetNotes)).Methods("GET")
	r.Handle("/api/gadgets/{id}/locations/{location}/devices/{device}/status", getMiddleware(handlers.Read, handlers.GetDevice)).Methods("GET")
	r.Handle("/api/gadgets/{id}/locations/{location}/devices/{device}/status", getMiddleware(handlers.Write, handlers.UpdateDevice)).Methods("POST")
	r.Handle("/api/gadgets/{id}/sources", getMiddleware(handlers.Read, handlers.GetDataPointSources)).Methods("GET")
	r.Handle("/api/gadgets/{id}/sources/{name}", getMiddleware(handlers.Read, handlers.GetDataPoints)).Methods("GET")
	r.Handle("/api/gadgets/{id}/sources/{name}/csv", getMiddleware(handlers.Read, handlers.GetDataPointsCSV)).Methods("GET")
	r.Handle("/api/beer/{name}", getMiddleware(handlers.Read, handlers.GetRecipe)).Methods("GET")
	r.Handle("/api/admin/clients", getMiddleware(handlers.Admin, handlers.GetClients)).Methods("GET")

	r.Handle("/css/{file}", http.FileServer(box.HTTPBox())).Methods("GET")
	r.Handle("/js/{file}", http.FileServer(box.HTTPBox())).Methods("GET")

	http.Handle(root, r)

	addr := fmt.Sprintf("%s:%s", iface, port)
	lg.Printf("listening on %s\n", addr)
	if keyPath == "" {
		lg.Println(http.ListenAndServe(addr, r))
	} else {
		lg.Println(http.ListenAndServeTLS(fmt.Sprintf("%s:443", iface), certPath, keyPath, r))
	}
}

func startHomeKit(db *bolt.DB, lg quimby.Logger) {
	key := os.Getenv("QUIMBY_HOMEKIT")
	if key == "" {
		lg.Println("QUIMBY_HOMEKIT not set, not starting homekit")
		return
	}
	hk := quimby.NewHomeKit(key, db)
	hk.Start()
}

//This is the endpoint that the gadgets report to. It is
//served on a separate port so it doesn't have to be exposed
//publicly if the main port is exposed.
func startInternal(iRoot string, db *bolt.DB, lg quimby.Logger, port string) {
	r := mux.NewRouter()
	r.Handle("/internal/updates", getMiddleware(handlers.Write, handlers.RelayMessage)).Methods("POST")
	r.Handle("/internal/gadgets/{id}/sources/{name}", getMiddleware(handlers.Write, handlers.AddDataPoint)).Methods("POST")

	chain := alice.New(handlers.Auth(), handlers.FetchGadget()).Then(r)

	http.Handle(iRoot, chain)
	a := fmt.Sprintf(":%s", port)
	lg.Printf("listening on %s", a)
	err := http.ListenAndServe(a, chain)
	if err != nil {
		log.Fatal(err)
	}
}
