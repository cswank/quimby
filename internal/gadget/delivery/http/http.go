package gadgethttp

import (
	"net/http"
	"strconv"

	rice "github.com/GeertJohan/go.rice"
	"github.com/cswank/quimby/internal/gadget"
	"github.com/cswank/quimby/internal/gadget/usecase"
	"github.com/cswank/quimby/internal/middleware"
	"github.com/cswank/quimby/internal/schema"
	"github.com/go-chi/chi"
)

func New(r chi.Router, box *rice.Box) {
	g := &GadgetHTTP{
		box:     box,
		usecase: usecase.New(),
	}

	r.Get("/", middleware.Handle(g.Redirect))
	r.Get("/gadgets", middleware.Handle(middleware.Render(g.GetAll)))
	r.Get("/gadgets/{id}", middleware.Handle(middleware.Render(g.Get)))
	r.Get("/static/*", middleware.Handle(g.Static()))

}

// GadgetHTTP renders html
type GadgetHTTP struct {
	usecase gadget.Usecase
	box     *rice.Box
}

type link struct {
	Name     string
	Link     string
	Selected string
	Children []link
}

type page struct {
	name        string
	Links       []link
	Scripts     []string
	Stylesheets []string
	template    string
}

func (p *page) Name() string {
	return p.name
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
	gadgets, err := g.usecase.GetAll()
	if err != nil {
		return nil, err
	}

	return &gadgetsPage{
		Gadgets: gadgets,
		page: page{
			name:     "Quimby",
			template: "gadgets.ghtml",
		},
	}, nil
}

type gadgetPage struct {
	page
	Gadget schema.Gadget
}

// Get shows a single gadget
func (g GadgetHTTP) Get(w http.ResponseWriter, req *http.Request) (middleware.Renderer, error) {
	id, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		return nil, err
	}

	gadget, err := g.usecase.Get(int(id))
	if err != nil {
		return nil, err
	}

	return &gadgetPage{
		Gadget: gadget,
		page: page{
			name:     gadget.Name,
			template: "gadget.ghtml",
		},
	}, nil
}

func (g GadgetHTTP) Static() middleware.Handler {
	s := http.FileServer(g.box.HTTPBox())
	return func(w http.ResponseWriter, req *http.Request) error {
		s.ServeHTTP(w, req)
		return nil
	}
}

// Redirect -> /gadgets
func (g GadgetHTTP) Redirect(w http.ResponseWriter, req *http.Request) error {
	http.Redirect(w, req, "/gadgets", http.StatusSeeOther)
	return nil
}
