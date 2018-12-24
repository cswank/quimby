package gadget

import "github.com/cswank/quimby/internal/schema"

// Repository stores gadgets
type Repository interface {
	GetAll() ([]schema.Gadget, error)
	Get(id string) (schema.Gadget, error)
}
