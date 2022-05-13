package repository

import (
	"github.com/asdine/storm"
	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/storage"
)

// Repo does database-y things.
type Repo struct {
	db *storm.DB
}

func New() *Repo {
	return &Repo{
		db: storage.Get(),
	}
}

func (r Repo) GetAll() ([]schema.Gadget, error) {
	var g []schema.Gadget
	return g, r.db.All(&g)
}

func (r Repo) Get(id int) (schema.Gadget, error) {
	var g schema.Gadget
	return g, r.db.One("ID", id, &g)
}

func (r Repo) Create(name, url string) (*schema.Gadget, error) {
	g := &schema.Gadget{Name: name, URL: url}
	return g, r.db.Save(g)
}

func (r Repo) Delete(id int) error {
	g := &schema.Gadget{ID: id}
	return r.db.DeleteStruct(g)
}

func (r Repo) Edit(g schema.Gadget) error {
	return r.db.Update(g)
}
