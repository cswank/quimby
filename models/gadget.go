package models

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
)

type GadgetHosts struct {
	Gadgets []Gadget `json:"gadgets"`
}

type Gadget struct {
	Name string `json:"name"`
	Host string `json:"host"`
	DB   *bolt.DB
}

func GetGadgets(db *bolt.DB) (*GadgetHosts, error) {
	gadgets := &GadgetHosts{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("gadgets"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var g Gadget
			if err := json.Unmarshal(v, &g); err != nil {
				return err
			}
			gadgets.Gadgets = append(gadgets.Gadgets, g)
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

func (g *Gadget) Status() (map[string]map[string]gogadgets.Value, error) {
	m := map[string]map[string]gogadgets.Value{}
	r, err := http.Get(fmt.Sprintf("%s/gadgets", g.Host))
	if err != nil {
		return m, err
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return m, dec.Decode(&m)
}
