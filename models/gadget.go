package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
)

type Gadget struct {
	Name string   `json:"name"`
	Host string   `json:"host"`
	DB   *bolt.DB `json:"-"`
}

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
		v := b.Get([]byte(g.Name))
		return json.Unmarshal(v, g)
	})
}

func (g *Gadget) Save() error {
	return g.DB.Update(func(tx *bolt.Tx) error {
		d, _ := json.Marshal(g)
		b := tx.Bucket([]byte("gadgets"))
		return b.Put([]byte(g.Name), d)
	})
}

func (g *Gadget) Delete() error {
	return g.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("gadgets"))
		return b.Delete([]byte(g.Name))
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

func (g *Gadget) Status(w io.Writer) error {
	r, err := http.Get(fmt.Sprintf("%s/gadgets", g.Host))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	_, err = io.Copy(w, r.Body)
	return err
}

func (g *Gadget) Register(addr string) (string, error) {
	if g.Host == "" {
		if err := g.Fetch(); err != nil {
			fmt.Println(g, err)
			return "", err
		}
	}
	a := map[string]string{"address": addr}
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.Encode(&a)
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
