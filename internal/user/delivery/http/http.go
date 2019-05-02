package userhttp

import (
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/cswank/quimby/internal/errors"
	"github.com/cswank/quimby/internal/middleware"
	"github.com/cswank/quimby/internal/templates"
	"github.com/cswank/quimby/internal/user"
	"github.com/cswank/quimby/internal/user/usecase"
	"github.com/go-chi/chi"
)

// userHTTP renders html
type userHTTP struct {
	usecase user.Usecase
	box     *rice.Box
	auth    *middleware.Auth
}

func Handle(r chi.Router, box *rice.Box) {
	u := &userHTTP{
		usecase: usecase.New(),
		box:     box,
		auth:    middleware.NewAuth(),
	}

	r.Route("/login", func(r chi.Router) {
		r.Get("/", middleware.Handle(middleware.Render(u.renderLogin)))
		r.Post("/", middleware.Handle(u.login))
	})

	r.Route("/logout", func(r chi.Router) {
		r.Get("/", middleware.Handle(middleware.Render(u.renderLogout)))
		r.Post("/", middleware.Handle(u.logout))
	})
}

func (u *userHTTP) renderLogin(w http.ResponseWriter, req *http.Request) (middleware.Renderer, error) {
	return &loginPage{
		Page:  templates.NewPage("Quimby", "login.ghtml"),
		Error: req.URL.Query().Get("error"),
	}, nil
}

func (u *userHTTP) renderLogout(w http.ResponseWriter, req *http.Request) (middleware.Renderer, error) {
	p := templates.NewPage("Quimby", "logout.ghtml")
	return &p, nil
}

func (u *userHTTP) login(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	username := req.Form.Get("username")
	pw := req.Form.Get("password")
	token := req.Form.Get("token")
	if err := u.usecase.Check(username, pw, token); err != nil {
		return errors.NewErrUnauthorized(err)
	}

	cookie, err := u.auth.GenerateCookie(username)
	if err != nil {
		return err
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, req, "/gadgets", http.StatusSeeOther)
	return nil
}

func (u *userHTTP) logout(w http.ResponseWriter, req *http.Request) error {
	cookie := &http.Cookie{
		Name:   "quimby",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, req, "/login", http.StatusSeeOther)
	return nil
}

type loginPage struct {
	templates.Page
	Error string
}
