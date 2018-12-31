package repository

import (
	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/user"
	"github.com/cswank/quimby/internal/user/repository"
	"golang.org/x/crypto/bcrypt"
)

// Usecase does nondatabase-y things
type Usecase struct {
	repo user.Repository
}

func New() *Usecase {
	return &Usecase{
		repo: repository.New(),
	}
}

func (u Usecase) GetAll() ([]schema.User, error) {
	return u.repo.GetAll()
}

func (u Usecase) Get(id int) (schema.User, error) {
	return u.repo.Get(id)
}

func (u Usecase) Create(name, pws string) (*schema.User, error) {
	pw, err := bcrypt.GenerateFromPassword([]byte(pws), 10)
	if err != nil {
		return nil, err
	}

	return u.repo.Create(name, pw)
}
