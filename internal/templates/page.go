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

func NewPage(name, template string) Page {
	return Page{
		name:     name,
		template: template,
	}
}

func (p *Page) Name() string {
	return p.name
}

func (p *Page) AddScripts(s []string) {
	p.Scripts = s
}

func (p *Page) AddLinks(l []Link) {
	p.Links = append(p.Links, l...)
}

func (p *Page) AddStylesheets(s []string) {
	p.Stylesheets = s
}

func (p *Page) Template() string {
	return p.template
}
