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
	index, login, logout, gadget, editGadget, admin *template.Template
)

func init() {
	parts := []string{"templates/head.html", "templates/base.html", "templates/navbar.html"}
	index = template.Must(template.ParseFiles(append(parts, "templates/index.html")...))
	gadget = template.Must(template.ParseFiles(append(parts, "templates/gadget.html", "templates/gadget.js", "templates/device.html")...))
	editGadget = template.Must(template.ParseFiles(append(parts, "templates/edit-gadget.html")...))
	admin = template.Must(template.ParseFiles(append(parts, "templates/admin.html")...))
	login = template.Must(template.ParseFiles(append(parts, "templates/login.html")...))
	logout = template.Must(template.ParseFiles(append(parts, "templates/logout.html")...))

}

type link struct {
	Name string
	Path string
}

type userPage struct {
	User  string
	Admin bool
	Links []link
}

type indexPage struct {
	userPage
	Gadgets []quimby.Gadget
}

type adminPage struct {
	userPage
	Gadgets []link
	Users   []link
}

type gadgetPage struct {
	userPage
	Gadget    *quimby.Gadget
	Websocket template.URL
	Locations map[string][]gogadgets.Message
	Error     string
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
			Links: []link{
				{"quimby", "/"},
			},
		},
	}
	index.ExecuteTemplate(w, "base", i)
}

func AdminPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	gadgets, err := quimby.GetGadgets(args.DB)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	links := make([]link, len(gadgets))
	for i, g := range gadgets {
		links[i] = link{Name: g.Name, Path: fmt.Sprintf("/admin/gadgets/%s", g.Id)}
	}

	users, err := quimby.GetUsers(args.DB)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userLinks := make([]link, len(users))
	for i, u := range users {
		userLinks[i] = link{Name: u.Username, Path: fmt.Sprintf("/admin/users/%s", u.Username)}
	}

	i := adminPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
			Links: []link{
				{"quimby", "/"},
				{"admin", "/admin.html"},
			},
		},
		Gadgets: links,
		Users:   userLinks,
	}
	admin.ExecuteTemplate(w, "base", i)
}

func displayValues(msg *gogadgets.Message) {
	if v, ok := msg.Value.Value.(float64); ok {
		msg.Value.Value = fmt.Sprintf("%.1f", v)
	}
}

func GadgetEditPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	g := gadgetPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
			Links: []link{
				{"quimby", "/"},
				{"admin", "/admin.html"},
				{args.Gadget.Name, fmt.Sprintf("/admin/gadgets/%s", args.Gadget.Id)},
			},
		},
		Gadget: args.Gadget,
	}
	editGadget.ExecuteTemplate(w, "base", g)
}

func GadgetForm(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		context.Set(req, "error", err)
		return
	}
	args := GetArgs(req)
	g := args.Gadget
	g.Host = req.PostFormValue("host")
	g.Name = req.PostFormValue("name")
	s := req.PostFormValue("disabled")

	d := s == "on"
	fmt.Println("dis", s, d)
	g.Disabled = d
	context.Set(req, "error", g.Save())
	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusMovedPermanently)
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
			Links: []link{
				{"quimby", "/"},
				{args.Gadget.Name, fmt.Sprintf("/admin/gadgets/%s", args.Gadget.Id)},
			},
		},
		Gadget:    args.Gadget,
		Websocket: template.URL(fmt.Sprintf("%s/api/gadgets/%s/websocket", u, args.Gadget.Id)),
		Locations: l,
	}
	gadget.ExecuteTemplate(w, "base", g)
}

func LoginPage(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	var p gadgetPage
	if q.Get("error") != "" {
		p.Error = "Invalid username or password"
	}
	login.ExecuteTemplate(w, "base", p)
}

func LogoutPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	p := userPage{
		User:  args.User.Username,
		Admin: Admin(args),
		Links: []link{
			{"quimby", "/"},
		},
	}
	logout.ExecuteTemplate(w, "base", p)
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
		w.Header().Set("Location", "/login.html?error=invalidlogin")
	} else {
		w.Header().Set("Location", "/index.html")
	}
	w.WriteHeader(http.StatusMovedPermanently)
}
