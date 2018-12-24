package gadgethttp

import (
	"net/http"

	"github.com/cswank/quimby/internal/gadget"
	"github.com/cswank/quimby/internal/gadget/repository"
	"github.com/cswank/quimby/internal/middleware"
	"github.com/cswank/quimby/internal/schema"
	"github.com/go-chi/chi"
)

func New(r chi.Router) {
	g := &GadgetHTTP{
		repo: repository.New(),
	}
	r.Get("/gadgets", middleware.Handle(middleware.Render(g.GetAll)))
}

// GadgetHTTP renders html
type GadgetHTTP struct {
	repo gadget.Repository
}

type link struct {
	Name     string
	Link     string
	Selected string
	Children []link
}

type page struct {
	Name        string
	Links       []link
	Scripts     []string
	Stylesheets []string
	template    string
}

func (p *page) AddScripts(s []string) {
	p.Scripts = s
}

func (p *page) AddStylesheets(s []string) {
	p.Stylesheets = s
}

func (p *page) Template() string {
	return p.template
}

type gadgetsPage struct {
	page
	Gadgets []schema.Gadget
}

// GetAll shows all the gadgets
func (g GadgetHTTP) GetAll(w http.ResponseWriter, req *http.Request) (middleware.Renderer, error) {
	gadgets, err := g.repo.GetAll()
	if err != nil {
		return nil, err
	}

	return &gadgetsPage{
		Gadgets: gadgets,
		page: page{
			template: "gadgets.ghtml",
		},
	}, nil
}
