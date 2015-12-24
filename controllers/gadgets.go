package controllers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/websocket"
)

var (
	beginning = time.Time{}
)

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

func AddNote(args *Args) error {
	var m map[string]string
	dec := json.NewDecoder(args.R.Body)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	n, ok := m["text"]
	if !ok {
		return fmt.Errorf("bad request")
	}
	return args.Gadget.AddNote(models.Note{Text: n, Author: args.User.Username})
}

func GetNotes(args *Args) error {
	notes, err := args.Gadget.GetNotes(nil, nil)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(args.W)
	return enc.Encode(notes)
}

func AddDataPoint(args *Args) error {
	var m map[string]float64
	dec := json.NewDecoder(args.R.Body)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	return args.Gadget.SaveDataPoint(args.Vars["name"], models.DataPoint{time.Now(), m["value"]})
}

func GetDataPoints(args *Args) error {
	start := beginning
	end := time.Now()
	if args.Args.Get("start") != "" {
		var err error
		start, err = time.Parse(time.RFC3339Nano, args.Args.Get("start"))
		if err != nil {
			return err
		}
	}

	if args.Args.Get("end") != "" {
		var err error
		end, err = time.Parse(time.RFC3339Nano, args.Args.Get("end"))
		if err != nil {
			return err
		}
	}
	points, err := args.Gadget.GetDataPoints(args.Vars["name"], start, end)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(args.W)
	return enc.Encode(points)
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

func SendMethod(args *Args) error {
	var m map[string][]string
	dec := json.NewDecoder(args.R.Body)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	return args.Gadget.Method(m["method"])
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
	if err := models.Register(*args.Gadget); err != nil {
		return err
	}
	ws := make(chan gogadgets.Message)
	q := make(chan bool)

	ch := make(chan gogadgets.Message)
	uuid := gogadgets.GetUUID()
	models.Clients.Add(args.Gadget.Host, uuid, ch)

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
			models.Clients.Delete(args.Gadget.Host, uuid)
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
	chs, ok := models.Clients.Get(m.Host)
	if !ok {
		return nil
	}
	for _, ch := range chs {
		ch <- m
	}
	return nil
}
