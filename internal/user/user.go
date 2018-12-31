package user

import "github.com/cswank/quimby/internal/schema"

// Repository stores users
type Repository interface {
	GetAll() ([]schema.User, error)
	Get(id int) (schema.User, error)
	Create(name string, pw, tfa []byte) (*schema.User, error)
}

// Usecase stores users
type Usecase interface {
	GetAll() ([]schema.User, error)
	Get(id int) (schema.User, error)
	Create(name string, pw []byte) (*schema.User, []byte, error)
}
