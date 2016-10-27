package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/cswank/quimby"
)

var (
	index, login *template.Template
)

func init() {
	parts := []string{"templates/head.html", "templates/base.html", "templates/navbar.html"}
	index = template.Must(template.ParseFiles(append(parts, "templates/index.html")...))
	login = template.Must(template.ParseFiles(append(parts, "templates/login.html")...))
}

type X struct {
	User  string
	Title string
	Body  string
}

func Index(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	x := X{Title: "my title", Body: "my body", User: args.User.Username}
	index.ExecuteTemplate(w, "base", x)
}

func LoginPage(w http.ResponseWriter, req *http.Request) {
	x := X{Title: "my title", Body: "login"}
	login.ExecuteTemplate(w, "base", x)
}

func LoginForm(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		fmt.Println("err")
		return
	}

	user := quimby.NewUser("", quimby.UserDB(DB), quimby.UserTFA(TFA))
	user.Username = req.PostFormValue("username")
	user.Password = req.PostFormValue("password")
	user.TFA = req.PostFormValue("tfa")
	if err := doLogin(user, w, req); err != nil {
		w.Header().Set("Location", "/login.html")
	} else {
		w.Header().Set("Location", "/index.html")
	}
	w.WriteHeader(http.StatusMovedPermanently)
}
