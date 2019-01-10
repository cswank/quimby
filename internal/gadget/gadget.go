package gadget

import "github.com/cswank/quimby/internal/schema"

// Repository stores gadgets
type Repository interface {
	GetAll() ([]schema.Gadget, error)
	Get(id int) (schema.Gadget, error)
	Create(name, url string) (*schema.Gadget, error)
	Delete(id int) error
}

// Usecase does non storage stuff with gadgets
type Usecase interface {
	GetAll() ([]schema.Gadget, error)
	Get(id int) (schema.Gadget, error)
	Create(name, url string) (*schema.Gadget, error)
}
