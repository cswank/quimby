package quimby

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/cswank/gogadgets"
)

type Client struct {
	addr  string
	Nodes []Gadget
	token string
	lock  sync.Mutex
}

func NewClient(addr string, opts ...Option) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	http.DefaultClient = &http.Client{Transport: tr}

	c := &Client{addr: fmt.Sprintf("%s/api/%%s", addr)}
	for _, opt := range opts {
		opt(c)
	}
	return c
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

	dec := json.NewDecoder(r.Body)
	return dec.Decode(&val)
}

func (c *Client) GetNodes() error {
	url := fmt.Sprintf(c.addr, "gadgets")
	if err := c.get(url, &c.Nodes); err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i, _ := range c.Nodes {
		wg.Add(1)
		go func(j int) {
			c.getNode(j, &wg)
		}(i)
	}
	wg.Wait()
	return nil
}

func (c *Client) getNode(i int, wg *sync.WaitGroup) {
	c.lock.Lock()
	g := c.Nodes[i]
	c.lock.Unlock()
	url := fmt.Sprintf(c.addr, fmt.Sprintf("gadgets/%s/values", g.Id))
	var m map[string]map[string]gogadgets.Value
	err := c.get(url, &m)
	if err == nil {
		c.lock.Lock()
		c.Nodes[i].Devices = m
		c.lock.Unlock()
	}
	wg.Done()
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
