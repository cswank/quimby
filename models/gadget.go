package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
)

type Gadget struct {
	Id   string   `json:"id"`
	Name string   `json:"name"`
	Host string   `json:"host"`
	DB   *bolt.DB `json:"-"`
}

var (
	NotFound = errors.New("not found")
)

func GetGadgets(db *bolt.DB) ([]Gadget, error) {
	gadgets := []Gadget{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("gadgets"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var g Gadget
			if err := json.Unmarshal(v, &g); err != nil {
				return err
			}
			gadgets = append(gadgets, g)
		}
		return nil
	})
	return gadgets, err
}

func (g *Gadget) Fetch() error {
	return g.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("gadgets"))
		v := b.Get([]byte(g.Id))
		if len(v) == 0 {
			return NotFound
		}
		return json.Unmarshal(v, g)
	})
}

func (g *Gadget) Save() error {
	if g.Id == "" {
		g.Id = gogadgets.GetUUID()
	}
	return g.DB.Update(func(tx *bolt.Tx) error {
		d, _ := json.Marshal(g)
		b := tx.Bucket([]byte("gadgets"))
		return b.Put([]byte(g.Id), d)
	})
}

func (g *Gadget) Delete() error {
	return g.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("gadgets"))
		return b.Delete([]byte(g.Id))
	})
}

func (g *Gadget) Update(cmd string) error {
	m := gogadgets.Message{
		UUID:   gogadgets.GetUUID(),
		Sender: "quimby",
		Type:   gogadgets.COMMAND,
		Body:   cmd,
	}
	return g.UpdateMessage(m)
}

func (g *Gadget) ReadDevice(w io.Writer, location, device string) error {
	r, err := http.Get(fmt.Sprintf("%s/gadgets/locations/%s/devices/%s/status", g.Host, location, device))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	_, err = io.Copy(w, r.Body)
	return err
}

func (g *Gadget) GetDevice(location string, name string) (gogadgets.Value, error) {
	m, err := g.GetValues()
	v, ok := m[location][name]
	if !ok {
		return v, fmt.Errorf("%s %s not found", location, name)
	}
	return v, err
}

func (g *Gadget) UpdateDevice(location string, name string, v gogadgets.Value) error {
	cmd := g.getCommand(location, name, v)
	return g.SendCommand(cmd)
}

func (g *Gadget) SendCommand(cmd string) error {
	m := gogadgets.Message{
		UUID:   gogadgets.GetUUID(),
		Sender: "quimby",
		Type:   gogadgets.COMMAND,
		Body:   cmd,
	}
	return g.UpdateMessage(m)
}

func (g *Gadget) getCommand(location string, name string, v gogadgets.Value) string {
	return fmt.Sprintf("%s %s %s", g.getVerb(v), location, name)
}

func (g *Gadget) getVerb(v gogadgets.Value) string {
	if v.Value == true {
		return "turn on"
	}
	return "turn off"
}

func (g *Gadget) UpdateMessage(m gogadgets.Message) error {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return err
	}
	r, err := http.Post(fmt.Sprintf("%s/gadgets", g.Host), "application/json", &buf)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from %s: %d", g.Host, r.StatusCode)
	}
	return nil
}

func (g *Gadget) Status() (map[string]gogadgets.Message, error) {
	var m map[string]gogadgets.Message
	r, err := http.Get(fmt.Sprintf("%s/gadgets", g.Host))

	if err != nil {
		return m, err
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return m, dec.Decode(&m)
}

func (g *Gadget) ReadStatus(w io.Writer) error {
	r, err := http.Get(fmt.Sprintf("%s/gadgets", g.Host))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	_, err = io.Copy(w, r.Body)
	return err
}

func (g *Gadget) GetValues() (map[string]map[string]gogadgets.Value, error) {
	r, err := http.Get(fmt.Sprintf("%s/gadgets/values", g.Host))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	var m map[string]map[string]gogadgets.Value
	dec := json.NewDecoder(r.Body)
	return m, dec.Decode(&m)
}

func (g *Gadget) ReadValues(w io.Writer) error {
	r, err := http.Get(fmt.Sprintf("%s/gadgets/values", g.Host))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	_, err = io.Copy(w, r.Body)
	return err
}

func (g *Gadget) Register(addr, token string) (string, error) {
	m := map[string]string{"address": addr, "token": token}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.Encode(&m)
	r, err := http.Post(fmt.Sprintf("%s/clients", g.Host), "application/json", buf)
	if err != nil {
		return "", err
	}
	r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response from %s: %d", g.Host, r.StatusCode)
	}
	return g.Host, nil
}
