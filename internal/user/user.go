package user

import "github.com/cswank/quimby/internal/schema"

// Repository stores users
type Repository interface {
	GetAll() ([]schema.User, error)
	Get(username string) (schema.User, error)
	Create(name string, pw, tfa []byte) (*schema.User, error)
}

// Usecase stores users
type Usecase interface {
	GetAll() ([]schema.User, error)
	Get(username string) (schema.User, error)
	Create(name, pw string) (*schema.User, []byte, error)
	Check(username, pw, token string) error
}
