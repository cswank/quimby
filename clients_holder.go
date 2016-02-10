package quimby

import (
	"encoding/json"
	"sync"

	"github.com/cswank/gogadgets"
)

//ClientHolder keeps track of the connected websocket clients.  It holds the chans
//that get written to when a Gogadgets system sends an update to Quimby.  That chan
//is used to relay the message the correct websocket.
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

func (c *ClientHolder) Add(host, key string, ch chan gogadgets.Message) {
	c.lock.Lock()
	chs, ok := c.clients[host]
	if !ok {
		chs = map[string](chan gogadgets.Message){}
	}
	chs[key] = ch
	c.clients[host] = chs
	c.lock.Unlock()
}

func (c *ClientHolder) Delete(host, uuid string) {
	c.lock.Lock()
	m, ok := c.clients[host]
	if ok {
		delete(m, uuid)
	}
	c.clients[host] = m
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
