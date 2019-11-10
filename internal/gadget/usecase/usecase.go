package usecase

import (
	"github.com/cswank/quimby/internal/gadget"
	"github.com/cswank/quimby/internal/gadget/repository"
	"github.com/cswank/quimby/internal/schema"
)

// Usecase does nondatabase-y things.
type Usecase struct {
	repo gadget.Repository
}

func New(opts ...func(*Usecase)) *Usecase {
	u := &Usecase{}

	for _, opt := range opts {
		opt(u)
	}

	if u.repo == nil {
		u.repo = repository.New()
	}

	return u
}

func Repo(r gadget.Repository) func(*Usecase) {
	return func(u *Usecase) {
		u.repo = r
	}
}

func (u Usecase) GetAll() ([]schema.Gadget, error) {
	return u.repo.GetAll()
}

func (u Usecase) Get(id int) (schema.Gadget, error) {
	g, err := u.repo.Get(id)
	if err != nil {
		return g, err
	}

	return g, g.Fetch()
}

func (u Usecase) Create(name, url string) (*schema.Gadget, error) {
	return u.repo.Create(name, url)
}
