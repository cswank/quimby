package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
)

var (
	addr         string
	clients      map[string]chan gogadgets.Message
	host         string
	hashKey      = []byte(os.Getenv("QUIMBY_HASH_KEY"))
	blockKey     = []byte(os.Getenv("QUIMBY_BLOCK_KEY"))
	SecureCookie = securecookie.New(hashKey, blockKey)
)

func init() {
	clients = map[string]chan gogadgets.Message{}
}

func Ping(args *Args) error {
	args.W.Header().Add(
		"Location",
		"/api/users/current",
	)
	return nil
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

//Registers with a gogadget instance and starts up
//a websocket.  It pushes new messages from the
//instance to the websocket and vice versa.
func Connect(args *Args) error {

	value := map[string]string{
		"user": args.User.Username,
	}

	encoded, _ := SecureCookie.Encode("quimby", value)
	cookie := &http.Cookie{
		Name:     "quimby",
		Value:    encoded,
		Path:     "/",
		HttpOnly: false,
	}

	h, err := args.Gadget.Register(getAddr(), cookie.String())
	if err != nil {
		return err
	}
	updates := make(chan gogadgets.Message)
	ws := make(chan gogadgets.Message)

	clients[h] = updates

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(args.W, args.R, nil)
	if err != nil {
		return err
	}

	go listen(conn, ws)

	for {
		select {
		case msg := <-ws:
			args.Gadget.UpdateMessage(msg)
		case msg := <-updates:
			sendSocketMessage(conn, msg)
		}
	}
	return nil
}

//Send a message via the web socket.
func sendSocketMessage(conn *websocket.Conn, m gogadgets.Message) {
	d, _ := json.Marshal(m)
	conn.WriteMessage(websocket.TextMessage, d)
}

func listen(conn *websocket.Conn, ch chan<- gogadgets.Message) {
	for {
		t, p, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if t == websocket.TextMessage {
			var m gogadgets.Message
			if err := json.Unmarshal(p, &m); err != nil {
				return
			}
			ch <- m
		} else if t == websocket.CloseMessage || t == -1 {
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
	ch, ok := clients[m.Host]
	if !ok {
		return nil
	}
	ch <- m
	return nil
}

func getAddr() string {
	host = os.Getenv("QUIMBY_HOST")
	if host == "" {
		log.Println("please set QUIMBY_HOST")
	}
	if addr == "" {
		addr = fmt.Sprintf("%s:%s/internal/updates", os.Getenv("QUIMBY_HOST"), os.Getenv("QUIMBY_INTERNAL_PORT"))
	}
	return addr
}
