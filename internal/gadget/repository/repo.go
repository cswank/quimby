package repository

import (
	"github.com/asdine/storm"
	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/storage"
)

type Repo struct {
	db *storm.DB
}

func New() *Repo {
	return &Repo{
		db: storage.Get(),
	}
}

func (r Repo) GetAll() ([]schema.Gadget, error) {
	return []schema.Gadget{
		{Name: "g 1", URL: "/gadgets/1234"},
		{Name: "g 2", URL: "/gadgets/5678"},
	}, nil
}

func (r Repo) Get(id string) (schema.Gadget, error) {
	return schema.Gadget{Name: "g 1", URL: "/gadgets/1234"}, nil
}
