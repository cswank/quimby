package templates

type Page struct {
	name        string
	Links       []Link
	Scripts     []string
	Stylesheets []string
	template    string
}

type Link struct {
	Name string
	Link string
}

func NewPage(name, template string, opts ...func(*Page)) Page {
	p := Page{
		name:     name,
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

func WithLinks(l []Link) func(*Page) {
	return func(p *Page) {
		p.Links = l
	}
}

func (p *Page) Name() string {
	return p.name
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
