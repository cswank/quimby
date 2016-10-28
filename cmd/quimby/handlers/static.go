package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby"
	"github.com/gorilla/context"
)

var (
	index, login, logout, gadget, admin *template.Template
)

func init() {
	parts := []string{"templates/head.html", "templates/base.html", "templates/navbar.html"}
	index = template.Must(template.ParseFiles(append(parts, "templates/index.html")...))
	gadget = template.Must(template.ParseFiles(append(parts, "templates/gadget.html", "templates/gadget.js", "templates/device.html")...))
	admin = template.Must(template.ParseFiles(append(parts, "templates/admin.html")...))
	login = template.Must(template.ParseFiles(append(parts, "templates/login.html")...))
	logout = template.Must(template.ParseFiles(append(parts, "templates/logout.html")...))

}

type userPage struct {
	User  string
	Admin bool
}

type indexPage struct {
	userPage
	Gadgets []quimby.Gadget
}

type gadgetPage struct {
	userPage
	Gadget    *quimby.Gadget
	Websocket template.URL
	Locations map[string][]gogadgets.Message
}

func IndexPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	g, err := quimby.GetGadgets(args.DB)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	i := indexPage{
		Gadgets: g,
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
		},
	}
	index.ExecuteTemplate(w, "base", i)
}

func AdminPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	i := indexPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
		},
	}
	admin.ExecuteTemplate(w, "base", i)
}

func displayValues(msg *gogadgets.Message) {
	if v, ok := msg.Value.Value.(float64); ok {
		msg.Value.Value = fmt.Sprintf("%.1f", v)
	}
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
		displayValues(&msg)
		msgs = append(msgs, msg)
		l[msg.Location] = msgs
	}

	u := "ws://localhost:8111"

	g := gadgetPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
		},
		Gadget:    args.Gadget,
		Websocket: template.URL(fmt.Sprintf("%s/api/gadgets/%s/websocket", u, args.Gadget.Id)),
		Locations: l,
	}
	gadget.ExecuteTemplate(w, "base", g)
}

func LoginPage(w http.ResponseWriter, req *http.Request) {
	login.ExecuteTemplate(w, "base", nil)
}

func LogoutPage(w http.ResponseWriter, req *http.Request) {
	logout.ExecuteTemplate(w, "base", nil)
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
