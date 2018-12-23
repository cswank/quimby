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
		{Name: "g 1"},
		{Name: "g 2"},
	}, nil
}
