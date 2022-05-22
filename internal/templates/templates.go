package templates

import (
	"database/sql/driver"
	"embed"
	"fmt"
	"html/template"

	"github.com/cswank/quimby/internal/schema"
)

var (
	//go:embed static/*
	Static embed.FS

	//go:embed templates/*
	tpls embed.FS

	templates map[string]tmpl

	deviceFuncs = template.FuncMap{
		"format": func(v driver.Value, decimals int) string {
			if v == nil {
				return ""
			}
			t := fmt.Sprintf("%%.%df", decimals)
			return fmt.Sprintf(t, v.(float64))
		},
		"command": func(devices map[string]map[string]schema.Message) map[string]schema.Message {
			out := map[string]schema.Message{}
			for location, statuses := range devices {
				for dev, status := range statuses {
					out[fmt.Sprintf("%s-%s", location, dev)] = status
				}
			}
			return out
		},
	}
)

type tmpl struct {
	template    *template.Template
	files       []string
	scripts     []string
	stylesheets []string
	funcs       template.FuncMap
	bare        bool
}

func Init() error {
	data := map[string]string{}
	files, err := tpls.ReadDir("templates")
	if err != nil {
		return err
	}

	for _, f := range files {
		d, err := tpls.ReadFile(fmt.Sprintf("templates/%s", f.Name()))
		if err != nil {
			return err
		}
		data[f.Name()] = string(d)
	}

	templates = map[string]tmpl{
		"login.ghtml":       {},
		"logout.ghtml":      {},
		"gadgets.ghtml":     {},
		"edit-method.ghtml": {files: []string{"edit-method.js"}},
		"gadget.ghtml":      {funcs: deviceFuncs, files: []string{"device.ghtml", "method.ghtml", "gadgets.js", "method.js"}, stylesheets: []string{"/static/switch.css"}},
	}

	base := []string{"head.ghtml", "base.ghtml", "navbar.ghtml", "menu-item.ghtml"}

	for key, val := range templates {
		t := template.New(key)
		if val.funcs != nil {
			t = t.Funcs(val.funcs)
		}
		var err error
		files := append([]string{key}, val.files...)
		files = append(files, base...)
		for _, f := range files {
			t, err = t.Parse(data[f])
			if err != nil {
				return err
			}
		}
		val.template = t
		templates[key] = val
	}

	return nil
}

func Get(k string) (*template.Template, []string, []string) {
	t := templates[k]
	return t.template, t.scripts, t.stylesheets
}
