package userhttp

import (
	"net/http"

	rice "github.com/GeertJohan/go.rice"
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

func Init(r chi.Router, box *rice.Box) {
	u := &userHTTP{
		usecase: usecase.New(),
		box:     box,
		auth:    middleware.NewAuth(),
	}

	r.Get("/login", middleware.Handle(middleware.Render(u.render("login.ghtml"))))
	r.Post("/login", middleware.Handle(u.login))
	r.Get("/logout", middleware.Handle(middleware.Render(u.render("logout.ghtml"))))
	r.Post("/logout", middleware.Handle(u.logout))
}

func (u *userHTTP) render(template string) middleware.RenderFunc {
	return func(w http.ResponseWriter, req *http.Request) (middleware.Renderer, error) {
		p := templates.NewPage("Quimby", template)
		return &p, nil
	}
}

func (u *userHTTP) login(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	username := req.Form.Get("username")
	pw := req.Form.Get("password")
	token := req.Form.Get("token")
	if err := u.usecase.Check(username, pw, token); err != nil {
		return err
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
