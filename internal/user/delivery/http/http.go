package userhttp

import (
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/cswank/quimby/internal/middleware"
	"github.com/cswank/quimby/internal/templates"
	"github.com/cswank/quimby/internal/user"
	"github.com/cswank/quimby/internal/user/repository"
	"github.com/go-chi/chi"
)

// userHTTP renders html
type userHTTP struct {
	repo user.Repository
	box  *rice.Box
}

func Init(r chi.Router, box *rice.Box) {
	u := &userHTTP{
		repo: repository.New(),
		box:  box,
	}

	r.Get("/login", middleware.Handle(middleware.Render(u.renderLogin)))
	r.Post("/login", middleware.Handle(u.login))
}

func (u *userHTTP) renderLogin(w http.ResponseWriter, req *http.Request) (middleware.Renderer, error) {
	p := templates.NewPage("login", "login.ghtml")
	return &p, nil
}

func (u *userHTTP) login(w http.ResponseWriter, req *http.Request) error {
	return nil
}
