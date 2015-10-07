package controllers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/websocket"
)

type ClientHolder struct {
	clients map[string]map[string](chan gogadgets.Message)
	lock    sync.Mutex
}

func (c *ClientHolder) Get(key string) (map[string](chan gogadgets.Message), bool) {
	c.lock.Lock()
	m, ok := c.clients[key]
	c.lock.Unlock()
	return m, ok
}

func (c *ClientHolder) Add(key string, chs map[string](chan gogadgets.Message)) {
	c.lock.Lock()
	c.clients[key] = chs
	c.lock.Unlock()
}

func (c *ClientHolder) MarshalJSON() ([]byte, error) {
	m := map[string]int{}
	c.lock.Lock()
	for k, v := range Clients.clients {
		m[k] = len(v)
	}
	c.lock.Unlock()
	return json.Marshal(m)
}

func NewClientHolder() *ClientHolder {
	return &ClientHolder{
		clients: make(map[string]map[string](chan gogadgets.Message)),
	}
}

func Ping(args *Args) error {
	args.W.Header().Add(
		"Location",
		"/api/users/current",
	)
	return nil
}

func GetClients(args *Args) error {
	d, err := json.Marshal(Clients)
	args.W.Write(d)
	return err
}

func GetGadgets(args *Args) error {
	g, err := models.GetGadgets(args.DB)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(args.W)
	return enc.Encode(g)
}

func GetGadget(args *Args) error {
	enc := json.NewEncoder(args.W)
	return enc.Encode(args.Gadget)
}

func GetUser(args *Args) error {
	enc := json.NewEncoder(args.W)
	return enc.Encode(args.User)
}

func DeleteGadget(args *Args) error {
	return args.Gadget.Delete()
}

func GetStatus(args *Args) error {
	return args.Gadget.ReadStatus(args.W)
}

func GetValues(args *Args) error {
	return args.Gadget.ReadValues(args.W)
}

func SendCommand(args *Args) error {
	var m map[string]string
	dec := json.NewDecoder(args.R.Body)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	return args.Gadget.Update(m["command"])
}

func AddGadget(args *Args) error {
	var g models.Gadget
	dec := json.NewDecoder(args.R.Body)
	err := dec.Decode(&g)
	if err != nil {
		return err
	}
	g.DB = args.DB
	err = g.Save()
	if err != nil {
		return err
	}

	u, err := url.Parse(fmt.Sprintf("/api/gadgets/%s", g.Name))
	if err != nil {
		return err
	}
	args.W.Header().Set("Location", u.String())
	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//Registers with a gogadget instance and starts up
//a websocket.  It pushes new messages from the
//instance to the websocket and vice versa.
func Connect(args *Args) error {
	token, err := generateToken(args.User)
	if err != nil {
		return err
	}

	h, err := args.Gadget.Register(getAddr(), token)
	if err != nil {
		return err
	}
	ch := make(chan gogadgets.Message)
	ws := make(chan gogadgets.Message)
	q := make(chan bool)

	chs, ok := Clients.Get(h)
	if !ok {
		chs = map[string](chan gogadgets.Message){}
	}

	uuid := gogadgets.GetUUID()
	chs[uuid] = ch
	Clients.Add(h, chs)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(args.W, args.R, nil)
	if err != nil {
		return err
	}

	go listen(conn, ws, q)

	for {
		select {
		case msg := <-ws:
			args.Gadget.UpdateMessage(msg)
		case msg := <-ch:
			sendSocketMessage(conn, msg)
		case <-q:
			m := Clients.clients[h]
			delete(m, uuid)
			return nil
		}
	}
	return nil
}

//Send a message via the web socket.
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

func GetDevice(args *Args) error {
	return args.Gadget.ReadDevice(args.W, args.Vars["location"], args.Vars["device"])
}

func UpdateDevice(args *Args) error {
	var v gogadgets.Value
	dec := json.NewDecoder(args.R.Body)
	if err := dec.Decode(&v); err != nil {
		return err
	}
	return args.Gadget.UpdateDevice(args.Vars["location"], args.Vars["device"], v)
}

func RelayMessage(args *Args) error {
	var m gogadgets.Message
	dec := json.NewDecoder(args.R.Body)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	chs, ok := Clients.clients[m.Host]
	if !ok {
		return nil
	}
	for _, ch := range chs {
		ch <- m
	}
	return nil
}

func getAddr() string {
	host = os.Getenv("QUIMBY_HOST")
	if host == "" {
		LG.Println("please set QUIMBY_HOST")
	}
	if addr == "" {
		addr = fmt.Sprintf("%s:%s/internal/updates", os.Getenv("QUIMBY_HOST"), os.Getenv("QUIMBY_INTERNAL_PORT"))
	}
	return addr
}
