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

func (r Repo) GetAll() ([]schema.User, error) {
	var g []schema.User
	return g, r.db.All(&g)
}

func (r Repo) Get(id int) (schema.User, error) {
	var g schema.User
	return g, r.db.One("ID", id, &g)
}

func (r Repo) Create(name string, pw, tfa []byte) (*schema.User, error) {
	u := &schema.User{Name: name, Password: pw, TFA: tfa}
	return u, r.db.Save(u)
}
