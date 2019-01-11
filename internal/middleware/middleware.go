package middleware

import (
	"log"
	"net/http"

	"github.com/cswank/quimby/internal/errors"
	"github.com/cswank/quimby/internal/templates"
)

type Handler func(http.ResponseWriter, *http.Request) error
type RenderFunc func(http.ResponseWriter, *http.Request) (Renderer, error)

func Handle(h Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := h(w, req)
		if err == nil {
			return
		}

		log.Println(err)
		if errors.IsUnauthorized(err) {
			http.Redirect(w, req, `/login?error="invalid login"`, http.StatusSeeOther)
		}
	}
}

func Render(r RenderFunc) func(w http.ResponseWriter, req *http.Request) error {
	return func(w http.ResponseWriter, req *http.Request) error {
		pg, err := r(w, req)
		if err != nil || pg == nil {
			return err
		}

		t, scripts, stylesheets := templates.Get(pg.Template())
		pg.AddScripts(scripts)
		pg.AddStylesheets(stylesheets)
		pg.AddLinks([]templates.Link{{Name: "logout", Link: "/logout"}})
		return t.ExecuteTemplate(w, "base", pg)
	}
}

// Renderer supplies the data needed to render html
type Renderer interface {
	Name() string
	AddScripts([]string)
	AddLinks([]templates.Link)
	AddStylesheets([]string)
	Template() string
}
