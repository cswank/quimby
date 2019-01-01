package templates

import (
	"database/sql/driver"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	rice "github.com/GeertJohan/go.rice"
)

var (
	templates map[string]tmpl

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

func Box(box *rice.Box) {
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
		"login.ghtml":   {},
		"logout.ghtml":  {},
		"gadgets.ghtml": {},
		"gadget.ghtml":  {files: []string{"device.ghtml"}, stylesheets: []string{"/static/switch.css"}},
	}

	base := []string{"head.ghtml", "base.ghtml", "navbar.ghtml", "menu-item.ghtml", "base.js"}

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

func getHTML(box *rice.Box) ([]string, error) {
	var html []string
	return html, box.Walk("/", func(pth string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(pth, ".ghtml") || strings.HasSuffix(pth, ".js") {
			if box.IsEmbedded() {
				pth = pth[1:] //workaround until https://github.com/GeertJohan/go.rice/issues/71 is fixed (which is probably never)
			}
			html = append(html, pth)
		}
		return nil
	})
}

func Get(k string) (*template.Template, []string, []string) {
	t := templates[k]
	return t.template, t.scripts, t.stylesheets
}
