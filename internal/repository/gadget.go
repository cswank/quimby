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

func (g Gadget) Update(id int, name, url string) error {
	return g.db.Update(&schema.Gadget{ID: id, Name: name, URL: url})
}

func (g Gadget) List() ([]schema.Gadget, error) {
	var gs []schema.Gadget
	return gs, g.db.All(&gs)
}
