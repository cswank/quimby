package gadget

import "github.com/cswank/quimby/internal/schema"

// Repository stores gadgets
type Repository interface {
	GetAll() ([]schema.Gadget, error)
}

// Usecase validates gadgets.
type Usecase interface {
	GetAll() ([]schema.Gadget, error)
}
