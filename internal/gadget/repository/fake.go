package repository

import (
	"github.com/cswank/quimby/internal/schema"
)

type Fake struct {
	DoGetAll func() ([]schema.Gadget, error)
	DoGet    func(id int) (schema.Gadget, error)
	DoCreate func(name, url string) (*schema.Gadget, error)
	DoDelete func(id int) error
}

func (f *Fake) GetAll() ([]schema.Gadget, error) {
	return f.DoGetAll()
}

func (f *Fake) Get(id int) (schema.Gadget, error) {
	return f.DoGet(id)
}

func (f *Fake) Create(name, url string) (*schema.Gadget, error) {
	return f.DoCreate(name, url)
}

func (f *Fake) Delete(id int) error {
	return f.DoDelete(id)
}
