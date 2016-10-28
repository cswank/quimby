package handlers

import (
	"html/template"
	"net/http"

	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby"
	"github.com/gorilla/context"
)

var (
	index, login, gadget *template.Template
)

func init() {
	parts := []string{"templates/head.html", "templates/base.html", "templates/navbar.html"}
	index = template.Must(template.ParseFiles(append(parts, "templates/index.html")...))
	gadget = template.Must(template.ParseFiles(append(parts, "templates/gadget.html", "templates/gadget.js")...))
	login = template.Must(template.ParseFiles(append(parts, "templates/login.html")...))
}

type indexPage struct {
	User    string
	Gadgets []quimby.Gadget
}

type gadgetPage struct {
	User      string
	Gadget    *quimby.Gadget
	Locations map[string][]gogadgets.Message
}

func Index(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	g, err := quimby.GetGadgets(args.DB)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	i := indexPage{Gadgets: g, User: args.User.Username}
	index.ExecuteTemplate(w, "base", i)
}

func GadgetPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	s, err := args.Gadget.Status()

	if err != nil {
		context.Set(req, "error", err)
		return
	}

	l := map[string][]gogadgets.Message{}
	for _, msg := range s {
		msgs, ok := l[msg.Location]
		if !ok {
			msgs = []gogadgets.Message{}
		}
		msgs = append(msgs, msg)
		l[msg.Location] = msgs
	}

	g := gadgetPage{
		User:      args.User.Username,
		Gadget:    args.Gadget,
		Locations: l,
	}
	gadget.ExecuteTemplate(w, "base", g)
}

func LoginPage(w http.ResponseWriter, req *http.Request) {
	login.ExecuteTemplate(w, "base", nil)
}

func LoginForm(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		context.Set(req, "error", err)
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
