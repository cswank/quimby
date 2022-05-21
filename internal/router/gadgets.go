package router

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/templates"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

// getAll shows all the gadgets
func (g *server) getAll(w http.ResponseWriter, req *http.Request) error {
	rand.Seed(time.Now().UnixNano())
	gadgets, err := g.gadgets.GetAll()
	if err != nil {
		return err
	}

	return render(templates.NewPage("Quimby", "gadgets.ghtml", templates.WithGadgets(gadgets...)), w, req)
}

// get shows a single gadget
func (g *server) get(w http.ResponseWriter, req *http.Request) error {
	gadget, err := g.gadget(req)
	if err != nil {
		return err
	}

	return render(templates.NewPage(
		gadget.Name,
		"gadget.ghtml",
		templates.WithWebsocket(fmt.Sprintf("wss://%s/gadgets/%d/websocket", g.cfg.Host, gadget.ID)),
		templates.WithGadget(gadget),
		templates.WithScripts([]string{"https://cdnjs.cloudflare.com/ajax/libs/underscore.js/1.9.1/underscore-min.js"}),
		templates.WithLinks([]templates.Link{{Name: "method", Link: fmt.Sprintf("/gadgets/%d/method", gadget.ID)}}),
	), w, req)
}

// method shows the method editor for a gadget
func (g *server) method(w http.ResponseWriter, req *http.Request) error {
	gadget, err := g.gadget(req)
	if err != nil {
		return err
	}

	return render(templates.NewPage(
		"Quimby",
		"edit-method.ghtml",
		templates.WithScripts([]string{"https://cdnjs.cloudflare.com/ajax/libs/underscore.js/1.9.1/underscore-min.js"}),
		templates.WithGadget(gadget),
	), w, req)
}

func (g *server) runMethod(w http.ResponseWriter, req *http.Request) error {
	gadget, err := g.gadget(req)
	if err != nil {
		return err
	}

	var m schema.Method
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		return err
	}

	msg := schema.Message{
		Type:   "method",
		Method: m,
	}
	return gadget.Send(msg)
}

// connect registers with a gogadget instance and starts up
// a websocket.  It pushes new messages from the
// instance to the websocket and vice versa.
func (g *server) connect(w http.ResponseWriter, req *http.Request) error {
	gadget, err := g.gadget(req)
	if err != nil {
		return err
	}

	if err := g.register(gadget); err != nil {
		return err
	}

	ws := make(chan schema.Message)
	q := make(chan bool)

	ch := make(chan schema.Message)
	uuid := schema.UUID()
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

func (g *server) register(gadget schema.Gadget) error {
	_, err := gadget.Register(g.cfg.InternalAddress, g.randString())
	return err
}

func (g *server) randString() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Send a message via the web socket.
func sendSocketMessage(conn *websocket.Conn, m schema.Message) {
	d, _ := json.Marshal(m)
	if err := conn.WriteMessage(websocket.TextMessage, d); err != nil {
		log.Println("unable to write to websocket", err)
	}
}

func listen(conn *websocket.Conn, ch chan<- schema.Message, q chan<- bool) {
	for {
		t, p, err := conn.ReadMessage()
		if err != nil {
			q <- true
			return
		}
		if t == websocket.TextMessage {
			var m schema.Message
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

func (g *server) gadget(req *http.Request) (schema.Gadget, error) {
	id, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		return schema.Gadget{}, err
	}

	return g.gadgets.Get(int(id))
}

func (g *server) static() handler {
	s := http.FileServer(http.FS(templates.Static))
	return func(w http.ResponseWriter, req *http.Request) error {
		s.ServeHTTP(w, req)
		if strings.Contains(req.URL.Path, ".css.map") {
			w.Header()["content-type"] = []string{"text/css"}
		}
		return nil
	}
}

// update is where the gadgets post their updates to the UI.
func (g server) update(w http.ResponseWriter, req *http.Request) error {
	var msg schema.Message
	if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
		return err
	}

	g.hc.Update(msg)
	g.clients.Update(msg)
	return nil
}

// redirect -> /gadgets
func (g server) redirect(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "/gadgets", http.StatusSeeOther)
}

func noQueries(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL = &url.URL{Path: req.URL.Path}
		h.ServeHTTP(w, req)
	})
}
