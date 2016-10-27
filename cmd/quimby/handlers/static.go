package handlers

import (
	"html/template"
	"net/http"
)

var (
	tmpl *template.Template
)

func init() {
	tmpl = template.Must(template.ParseFiles("templates/index.html"))
}

type X struct {
	User  string
	Title string
	Body  string
}

func Index(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	x := X{Title: "my title", Body: "my body", User: args.User.Username}
	tmpl.Execute(w, x)
}
