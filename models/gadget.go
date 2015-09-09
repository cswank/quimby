package models

import (
	"encoding/json"

	"github.com/boltdb/bolt"
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
