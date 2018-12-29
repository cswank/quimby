package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cswank/gogadgets"
)

// Gadget represents a gadget
type Gadget struct {
	ID     int    `storm:"id,increment"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	status map[string]map[string]gogadgets.Message
}

func (g *Gadget) Status() map[string]map[string]gogadgets.Message {
	return g.status
}

func (g *Gadget) Register(addr, token string) (string, error) {
	m := map[string]string{"address": addr, "token": token}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(&m)
	if err != nil {
		return "", err
	}

	r, err := http.Post(fmt.Sprintf("%s/clients", g.URL), "application/json", buf)
	if err != nil {
		return "", err
	}

	r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response from %s: %d", g.URL, r.StatusCode)
	}

	return g.URL, nil
}

func (g *Gadget) Command(msg gogadgets.Message) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return err
	}

}

// Fetch queries the gadget to get its current status
func (g *Gadget) Fetch() error {
	resp, err := http.Get(fmt.Sprintf("%s/gadgets", g.URL))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	var m map[string]gogadgets.Message
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return err
	}

	status := map[string]map[string]gogadgets.Message{}

	for _, v := range m {
		if v.Name == "" || v.Location == "" {
			continue
		}

		l, ok := status[v.Location]
		if !ok {
			l = map[string]gogadgets.Message{}
		}

		l[v.Name] = v
		status[v.Location] = l
	}

	g.status = status
	return nil
}
