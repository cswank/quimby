package quimby

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/cswank/gogadgets"
	"github.com/gorilla/websocket"
)

type Node struct {
	Gadget
	Devices map[string]map[string]gogadgets.Message
}

type Nodes []Node

func (n Nodes) Less(i, j int) bool {
	return n[i].Name < n[j].Name
}

func (n Nodes) Len() int {
	return len(n)
}

func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

type Client struct {
	addr        string
	Nodes       Nodes
	Out         chan gogadgets.Message
	token       string
	lock        sync.Mutex
	ws          *websocket.Conn
	disconnect  chan bool
	connectedTo int
}

func NewClient(addr string, opts ...Option) (*Client, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		http.DefaultClient = &http.Client{Transport: tr}
	}

	c := &Client{addr: fmt.Sprintf("%s/api/%%s", addr)}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

func (c *Client) get(url string, val interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.token)
	req.Header.Add("Accept", "application/json")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return dec.Decode(&val)
}

func (c *Client) GetNodes() error {
	url := fmt.Sprintf(c.addr, "gadgets")
	if err := c.get(url, &c.Nodes); err != nil {
		return err
	}

	sort.Sort(c.Nodes)

	var wg sync.WaitGroup
	for i := range c.Nodes {
		wg.Add(1)
		go func(j int) {
			c.fetchNode(j, &wg)
		}(i)
	}
	wg.Wait()
	return nil
}

func (c *Client) fetchNode(i int, wg *sync.WaitGroup) {
	c.lock.Lock()
	n := c.Nodes[i]
	c.lock.Unlock()
	url := fmt.Sprintf(c.addr, fmt.Sprintf("gadgets/%s/status", n.Id))
	var m map[string]gogadgets.Message
	err := c.get(url, &m)
	if err == nil {
		c.lock.Lock()
		c.Nodes[i].Devices = c.getDevices(m)
		c.lock.Unlock()
	}
	wg.Done()
}

func (c *Client) getDevices(i map[string]gogadgets.Message) map[string]map[string]gogadgets.Message {
	m := map[string]map[string]gogadgets.Message{}
	for _, v := range i {
		a, ok := m[v.Location]
		if !ok {
			a = map[string]gogadgets.Message{}
		}
		a[v.Name] = v
		m[v.Location] = a
	}
	return m
}

func (c *Client) Login(username, password string) (string, error) {
	url := fmt.Sprintf(c.addr, "login?auth=jwt")
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	usr := &User{
		Username: username,
		Password: password,
	}
	enc.Encode(usr)

	r, err := http.Post(url, "application/json", &buf)
	if err != nil {
		return "", err
	}
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("couldn't log in: %d", r.StatusCode)
	}
	c.token = r.Header.Get("Authorization")
	return c.token, nil
}

func (c *Client) Disconnect() {
	c.disconnect <- true
}

func (c *Client) Connect(i int, cb func(gogadgets.Message)) error {
	c.disconnect = make(chan bool)
	c.connectedTo = i
	g := c.Nodes[i]

	a := fmt.Sprintf(c.addr, fmt.Sprintf("gadgets/%s/websocket", g.Id))
	u, err := url.Parse(a)
	if err != nil {
		return err
	}

	if u.Scheme == "https" {
		a = strings.Replace(a, "https", "wss", -1)
	} else {
		a = strings.Replace(a, "http", "ws", -1)
	}

	h := http.Header{"Origin": {a}, "Authorization": {c.token}}
	dialer := websocket.Dialer{
		Subprotocols:    []string{"p1", "p2"},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	c.ws, _, err = dialer.Dial(a, h)
	if err != nil {
		return err
	}
	c.Out = make(chan gogadgets.Message)
	ws := make(chan gogadgets.Message)
	go c.listen(ws)
	go func() {
		for {
			select {
			case msg := <-c.Out:
				d, _ := json.Marshal(msg)
				c.ws.WriteMessage(websocket.TextMessage, d)
			case msg := <-ws:
				cb(msg)
			case <-c.disconnect:
				c.ws.Close()
				return
			}
		}
	}()
	return nil
}

func (c *Client) listen(ch chan gogadgets.Message) {
	for {
		var msg gogadgets.Message
		if err := c.ws.ReadJSON(&msg); err != nil {
			return
		}
		c.lock.Lock()
		c.Nodes[c.connectedTo].Devices[msg.Location][msg.Name] = msg
		c.lock.Unlock()
		ch <- msg
	}
}
