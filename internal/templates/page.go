package templates

import "github.com/cswank/quimby/internal/schema"

type Page struct {
	Name        string
	Links       []Link
	Scripts     []string
	Stylesheets []string
	template    string
	Gadgets     []schema.Gadget
	Gadget      schema.Gadget
	Websocket   string
	Error       string
}

type Link struct {
	Name string
	Link string
}

func NewPage(name, template string, opts ...func(*Page)) Page {
	p := Page{
		Name:     name,
		template: template,
	}

	for _, opt := range opts {
		opt(&p)
	}

	return p
}

func WithScripts(s []string) func(*Page) {
	return func(p *Page) {
		p.Scripts = s
	}
}

func WithGadgets(g []schema.Gadget) func(*Page) {
	return func(p *Page) {
		p.Gadgets = g
	}
}

func WithWebsocket(s string) func(*Page) {
	return func(p *Page) {
		p.Websocket = s
	}
}

func WithGadget(g schema.Gadget) func(*Page) {
	return func(p *Page) {
		p.Gadget = g
	}
}

func WithLinks(l []Link) func(*Page) {
	return func(p *Page) {
		p.Links = l
	}
}

func (p *Page) AddScripts(s []string) {
	p.Scripts = append(p.Scripts, s...)
}

func (p *Page) AddLinks(l []Link) {
	p.Links = append(p.Links, l...)
}

func (p *Page) AddStylesheets(s []string) {
	p.Stylesheets = append(p.Stylesheets, s...)
}

func (p *Page) Template() string {
	return p.template
}
