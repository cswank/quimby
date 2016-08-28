package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby"
	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
)

var (
	beginning = time.Time{}
)

func Ping(w http.ResponseWriter, req *http.Request) {
	w.Header().Add(
		"Location",
		"/api/currentuser",
	)
}

func GetGadgets(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	g, err := quimby.GetGadgets(args.DB)
	if err != nil {
		context.Set(req, "error", err)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(g)
}

func GetGadget(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	args := GetArgs(req)
	enc.Encode(args.Gadget)
}

func GetUsers(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	users, err := quimby.GetUsers(args.DB)
	if err != nil {
		context.Set(req, "error", err)
		return // err
	}

	enc := json.NewEncoder(w)
	enc.Encode(users)
}

func GetCurrentUser(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	args := GetArgs(req)
	enc.Encode(args.User)
}

func GetUser(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	u := quimby.User{
		DB:       args.DB,
		Username: args.Vars["username"],
	}
	err := u.Fetch()
	if err != nil {
		context.Set(req, "error", err)
		return // err
	}
	u.HashedPassword = []byte{}
	enc := json.NewEncoder(w)
	enc.Encode(u)
}

func DeleteGadget(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	args.Gadget.Delete()
}

func GetStatus(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	args.Gadget.ReadStatus(w)
}

func AddNote(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	var m map[string]string
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		context.Set(req, "error", err)
		return // err
	}
	n, ok := m["text"]
	if ok {
		args.Gadget.AddNote(quimby.Note{Text: n, Author: args.User.Username})
	}
}

func GetNotes(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	notes, err := args.Gadget.GetNotes(nil, nil)
	if err != nil {
		context.Set(req, "error", err)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(notes)
}

func AddDataPoint(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	var m map[string]float64
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		context.Set(req, "error", err)
		return
	}
	args.Gadget.SaveDataPoint(args.Vars["name"], quimby.DataPoint{time.Now(), m["value"]})
}

func GetDataPointsCSV(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	points, err := getDataPoints(w, args)
	if err != nil {
		context.Set(req, "error", err)
		return
	}
	getCSV(w, points, args.Vars["name"])
}

func GetDataPoints(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	points, err := getDataPoints(w, args)
	if err != nil {
		context.Set(req, "error", err)
		return
	}
	if req.Header.Get("accept") == "application/csv" {
		getCSV(w, points, args.Vars["name"])

	} else {
		enc := json.NewEncoder(w)
		enc.Encode(points)
	}
}

func getDataPoints(w http.ResponseWriter, args *Args) ([]quimby.DataPoint, error) {
	start := beginning
	end := time.Now()
	var d time.Duration
	s := args.Args.Get("summarize")
	if s != "" {
		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			d = time.Duration(i) * time.Minute
		}
	}
	if args.Args.Get("start") != "" {
		var err error
		start, err = time.Parse(time.RFC3339Nano, args.Args.Get("start"))
		if err != nil {
			return nil, err
		}
	}

	if args.Args.Get("end") != "" {
		var err error
		end, err = time.Parse(time.RFC3339Nano, args.Args.Get("end"))
		if err != nil {
			return nil, err
		}
	}
	return args.Gadget.GetDataPoints(args.Vars["name"], start, end, d)
}

func getCSV(w http.ResponseWriter, points []quimby.DataPoint, name string) {
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", name))
	w.Header().Set("Content-Type", "application/csv")
	wr := csv.NewWriter(w)
	wr.Write([]string{"time", "value"})
	for _, p := range points {
		wr.Write([]string{
			p.Time.Format(time.RFC3339Nano),
			fmt.Sprintf("%f", p.Value),
		})
	}
	wr.Flush()
}

func GetUpdates(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	args.Gadget.ReadValues(w)
}

func SendCommand(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	var m map[string]string
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		context.Set(req, "error", err)
		return
	}
	args.Gadget.Update(m["command"])
}

func SendMethod(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	var m map[string][]string
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		context.Set(req, "error", err)
		return
	}
	args.Gadget.Method(m["method"])
}

func AddGadget(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	var g quimby.Gadget
	dec := json.NewDecoder(req.Body)
	err := dec.Decode(&g)
	if err != nil {
		context.Set(req, "error", err)
		return // err
	}

	g.DB = args.DB
	err = g.Save()
	if err != nil {
		context.Set(req, "error", err)
		return // err
	}

	u, err := url.Parse(fmt.Sprintf("/api/gadgets/%s", g.Id))
	if err != nil {
		context.Set(req, "error", err)
		return //err
	}
	w.Header().Set("Location", u.String())
}

func UpdateGadget(w http.ResponseWriter, req *http.Request) {

	args := GetArgs(req)
	var g quimby.Gadget
	dec := json.NewDecoder(req.Body)
	err := dec.Decode(&g)
	if err != nil {
		context.Set(req, "error", err)
		return // err
	}

	g.DB = args.DB
	g.Id = args.Vars["id"]
	err = g.Save()
	if err != nil {
		context.Set(req, "error", err)
		return // err
	}

	u, err := url.Parse(fmt.Sprintf("/api/gadgets/%s", g.Id))
	if err != nil {
		context.Set(req, "error", err)
		return //err
	}
	w.Header().Set("Location", u.String())
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
func Connect(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	if err := quimby.Register(*args.Gadget); err != nil {
		context.Set(req, "error", err)
		return // err
	}
	ws := make(chan gogadgets.Message)
	q := make(chan bool)

	ch := make(chan gogadgets.Message)
	uuid := gogadgets.GetUUID()
	quimby.Clients.Add(args.Gadget.Host, uuid, ch)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		context.Set(req, "error", err)
		return // err
	}

	go listen(conn, ws, q)

	for {
		select {
		case msg := <-ws:
			args.Gadget.UpdateMessage(msg)
		case msg := <-ch:
			sendSocketMessage(conn, msg)
		case <-q:
			quimby.Clients.Delete(args.Gadget.Host, uuid)
			return // nil
		}
	}
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

func GetDevice(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	args.Gadget.ReadDevice(w, args.Vars["location"], args.Vars["device"])
}

func UpdateDevice(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	var v gogadgets.Value
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&v); err != nil {
		context.Set(req, "error", err)
		return
	}
	args.Gadget.UpdateDevice(args.Vars["location"], args.Vars["device"], v)
}

func RelayMessage(w http.ResponseWriter, req *http.Request) {
	var m gogadgets.Message
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		context.Set(req, "error", err)
		return
	}
	chs, ok := quimby.Clients.Get(m.Host)
	if !ok {
		return
	}
	for _, ch := range chs {
		ch <- m
	}
}
