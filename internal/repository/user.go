package repository

import (
	"github.com/asdine/storm"
	"github.com/cswank/quimby/internal/schema"
)

// User does database-y things.
type User struct {
	db *storm.DB
}

func (u User) GetAll() ([]schema.User, error) {
	var out []schema.User
	return out, u.db.All(&out)
}

func (u User) Get(username string) (schema.User, error) {
	var out schema.User
	return out, u.db.One("Name", username, &out)
}

func (u User) Create(name string, pw, tfa []byte) (*schema.User, error) {
	out := &schema.User{Name: name, Password: pw, TFA: tfa}
	return out, u.db.Save(out)
}

func (u User) Delete(name string) error {
	out, err := u.Get(name)
	if err != nil {
		return err
	}
	return u.db.DeleteStruct(&out)
}
