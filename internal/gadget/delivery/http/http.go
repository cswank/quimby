package gadgethttp

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby/internal/clients"
	"github.com/cswank/quimby/internal/gadget"
	"github.com/cswank/quimby/internal/gadget/usecase"
	"github.com/cswank/quimby/internal/middleware"
	"github.com/cswank/quimby/internal/schema"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func Init(pub, priv chi.Router, box *rice.Box) {
	g := &GadgetHTTP{
		box:     box,
		usecase: usecase.New(),
		clients: clients.New(),
	}

	pub.Get("/", middleware.Handle(g.Redirect))
	pub.Get("/gadgets", middleware.Handle(middleware.Render(g.GetAll)))
	pub.Get("/gadgets/{id}", middleware.Handle(middleware.Render(g.Get)))
	pub.Get("/gadgets/{id}/websocket", middleware.Handle(g.Connect))
	pub.Get("/static/*", middleware.Handle(g.Static()))

	priv.Post("/status", middleware.Handle(g.Update))
}

// GadgetHTTP renders html
type GadgetHTTP struct {
	usecase     gadget.Usecase
	box         *rice.Box
	clients     *clients.Clients
	internalURL string
}

// GetAll shows all the gadgets
func (g *GadgetHTTP) GetAll(w http.ResponseWriter, req *http.Request) (middleware.Renderer, error) {
	rand.Seed(time.Now().UnixNano())
	gadgets, err := g.usecase.GetAll()
	if err != nil {
		return nil, err
	}

	return &gadgetsPage{
		Gadgets: gadgets,
		page: page{
			name:     "Quimby",
			template: "gadgets.ghtml",
		},
	}, nil
}

// Get shows a single gadget
func (g *GadgetHTTP) Get(w http.ResponseWriter, req *http.Request) (middleware.Renderer, error) {
	gadget, err := g.gadget(req)
	if err != nil {
		return nil, err
	}

	return &gadgetPage{
		Gadget:    gadget,
		Websocket: fmt.Sprintf("wss://localhost:3333/gadgets/%d/websocket", gadget.ID),
		page: page{
			name:     gadget.Name,
			template: "gadget.ghtml",
		},
	}, nil
}

// Connect registers with a gogadget instance and starts up
// a websocket.  It pushes new messages from the
// instance to the websocket and vice versa.
func (g *GadgetHTTP) Connect(w http.ResponseWriter, req *http.Request) error {
	gadget, err := g.gadget(req)
	fmt.Println("got here")

	if err != nil {
		return err
	}

	if err := g.register(gadget); err != nil {
		return err
	}

	ws := make(chan gogadgets.Message)
	q := make(chan bool)

	ch := make(chan gogadgets.Message)
	uuid := gogadgets.GetUUID()
	g.clients.Add(gadget.URL, uuid, ch)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return err
	}

	go listen(conn, ws, q)

	for {
		select {
		case msg := <-ws: // user is sending a command
			if err := gadget.Send(msg); err != nil {
				return err
			}
		case msg := <-ch: // gadget is sending an update to all those that care.
			sendSocketMessage(conn, msg)
		case <-q: // user has left the page where the websocket lived.
			g.clients.Delete(gadget.URL, uuid)
			return nil
		}
	}
}

func (g *GadgetHTTP) register(gadget schema.Gadget) error {
	_, err := gadget.Register(g.internalURL, g.randString())
	return err
}

func (g *GadgetHTTP) randString() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Send a message via the web socket.
func sendSocketMessage(conn *websocket.Conn, m gogadgets.Message) {
	d, _ := json.Marshal(m)
	conn.WriteMessage(websocket.TextMessage, d)
}

func listen(conn *websocket.Conn, ch chan<- gogadgets.Message, q chan<- bool) {
	for {
		t, p, err := conn.ReadMessage()
		if err != nil {
			q <- true
			return
		}
		if t == websocket.TextMessage {
			var m gogadgets.Message
			if err := json.Unmarshal(p, &m); err != nil {
				return
			}
			ch <- m
		} else if t == websocket.CloseMessage || t == -1 {
			q <- true
			return
		}
	}
}

func (g *GadgetHTTP) gadget(req *http.Request) (schema.Gadget, error) {
	id, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		return schema.Gadget{}, err
	}

	return g.usecase.Get(int(id))
}

func (g *GadgetHTTP) Static() middleware.Handler {
	s := http.FileServer(g.box.HTTPBox())
	return func(w http.ResponseWriter, req *http.Request) error {
		s.ServeHTTP(w, req)
		return nil
	}
}

// Update is where the gadgets post their updates to the UI.
func (g GadgetHTTP) Update(w http.ResponseWriter, req *http.Request) error {
	var msg gogadgets.Message
	if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
		return err
	}

	g.clients.Update(msg)
	return nil
}

// Redirect -> /gadgets
func (g GadgetHTTP) Redirect(w http.ResponseWriter, req *http.Request) error {
	http.Redirect(w, req, "/gadgets", http.StatusSeeOther)
	return nil
}

type link struct {
	Name     string
	Link     string
	Selected string
	Children []link
}

type page struct {
	name        string
	Links       []link
	Scripts     []string
	Stylesheets []string
	template    string
}

func (p *page) Name() string {
	return p.name
}

func (p *page) AddScripts(s []string) {
	p.Scripts = s
}

func (p *page) AddStylesheets(s []string) {
	p.Stylesheets = s
}

func (p *page) Template() string {
	return p.template
}

type gadgetsPage struct {
	page
	Gadgets []schema.Gadget
}

type gadgetPage struct {
	page
	Gadget    schema.Gadget
	Websocket string
}
