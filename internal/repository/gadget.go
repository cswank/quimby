package repository

import (
	"github.com/asdine/storm"
	"github.com/cswank/quimby/internal/schema"
)

// Gadget does database-y things.
type Gadget struct {
	db *storm.DB
}

func newGadget(db *storm.DB) *Gadget {
	return &Gadget{
		db: db,
	}
}

func (g Gadget) GetAll() ([]schema.Gadget, error) {
	var out []schema.Gadget
	return out, g.db.All(&out)
}

func (g Gadget) Get(id int) (schema.Gadget, error) {
	var out schema.Gadget
	return out, g.db.One("ID", id, &out)
}

func (g Gadget) Create(name, url string) (*schema.Gadget, error) {
	out := &schema.Gadget{Name: name, URL: url}
	return out, g.db.Save(out)
}

func (g Gadget) Delete(id int) error {
	out := &schema.Gadget{ID: id}
	return g.db.DeleteStruct(out)
}

func (r Gadget) Edit(id int) error {
	// g := &schema.Gadget{ID: id}
	// return r.db.DeleteStruct(g)
	return nil
}

func (r Gadget) List() ([]schema.Gadget, error) {
	//g := &schema.Gadget{ID: id}
	//return r.db.DeleteStruct(g)
	return nil, nil
}
