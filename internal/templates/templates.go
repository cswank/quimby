package templates

import (
	"database/sql/driver"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/gobuffalo/packr"
)

var (
	templates   map[string]tmpl
	deviceFuncs = template.FuncMap{
		"truncate": func(text string) string {
			if len(text) <= 10 {
				return text
			}
			return fmt.Sprintf("%s...", text[:10])
		},
		"format": func(v driver.Value, decimals int) string {
			if v == nil {
				return ""
			}
			t := fmt.Sprintf("%%.%df", decimals)
			return fmt.Sprintf(t, v.(float64))
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

func Init(box packr.Box) {
	data := map[string]string{}
	html, err := getHTML(box)
	if err != nil {
		log.Fatal(err)
	}

	for _, pth := range html {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(pth, err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"gadgets.html": {},
	}

	base := []string{"head.html", "base.html", "navbar.html"}

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
				log.Fatal(err)
			}
		}
		val.template = t
		templates[key] = val
	}
}

func getHTML(box packr.Box) ([]string, error) {
	var html []string
	box.Walk(func(pth string, f packr.File) error {
		info, err := f.FileInfo()
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(pth, ".html") || strings.HasSuffix(pth, ".js") {
			if box.IsEmbedded() {
				pth = pth[1:] //workaround until https://github.com/GeertJohan/go.rice/issues/71 is fixed (which is probably never)
			}
			html = append(html, pth)
		}
		return nil
	})
	return html
}

func Get(k string) (*template.Template, []string, []string) {
	t := templates[k]
	return t.template, t.scripts, t.stylesheets
}