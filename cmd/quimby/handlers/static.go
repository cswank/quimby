package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/GeertJohan/go.rice"
	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby"
	"github.com/gorilla/context"
)

var (
	ErrPasswordsDoNotMatch = errors.New("passwords do not match")
	templates              map[string]tmpl
)

type tmpl struct {
	template *template.Template
	files    []string
}

func Init(box *rice.Box) {
	data := map[string]string{}
	for _, pth := range []string{"head.html", "base.html", "navbar.html", "index.html", "gadget.html", "gadget.js", "device.html", "edit-gadget.html", "edit-user.html", "edit-user.js", "delete.html", "password.html", "new-user.html", "qr-code.html", "admin.html", "login.html", "logout.html"} {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"index.html":       {files: []string{"index.html"}},
		"gadget.html":      {files: []string{"gadget.html", "gadget.js", "device.html"}},
		"edit-gadget.html": {files: []string{"edit-gadget.html"}},
		"edit-user.html":   {files: []string{"edit-user.html", "edit-user.js"}},
		"delete.html":      {files: []string{"delete.html", "edit-user.js"}},
		"password.html":    {files: []string{"password.html", "edit-user.js"}},
		"new-user.html":    {files: []string{"new-user.html", "edit-user.js"}},
		"qr-code.html":     {files: []string{"qr-code.html"}},
		"admin.html":       {files: []string{"admin.html"}},
		"login.html":       {files: []string{"login.html"}},
		"logout.html":      {files: []string{"logout.html"}},
	}

	base := []string{"head.html", "base.html", "navbar.html"}

	for key, val := range templates {
		t := template.New(key)
		var err error
		for _, f := range append(val.files, base...) {
			t, err = t.Parse(data[f])
			if err != nil {
				log.Fatal(err)
			}
		}
		val.template = t
		templates[key] = val
	}
}

type link struct {
	Name string
	Path string
}

type action struct {
	Name   string
	URI    template.URL
	Method string
}

type userPage struct {
	User  string
	Admin bool
	Links []link
}

type editUserPage struct {
	userPage
	EditUser    *quimby.User
	Permissions []string
	Actions     []action
	End         int
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
	templates["index.html"].template.ExecuteTemplate(w, "base", i)
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
	templates["admin.html"].template.ExecuteTemplate(w, "base", i)
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
	templates["edit-gadget.html"].template.ExecuteTemplate(w, "base", g)
}

func UserEditPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	username := args.Vars["username"]
	page := editUserPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
		},
		Permissions: []string{"read", "write", "admin"},
	}
	if username == "new-user" {
		page.Links = []link{
			{"quimby", "/"},
			{"admin", "/admin.html"},
			{"new user", "/admin/users/new-user"},
		}
		page.Actions = []action{
			{Name: "cancel", URI: template.URL("/admin.html"), Method: "get"},
		}
		page.End = 0
	} else {
		u := quimby.NewUser(username, quimby.UserDB(args.DB))
		if err := u.Fetch(); err != nil {
			context.Set(req, "error", err)
			return
		}
		page.EditUser = u
		page.Links = []link{
			{"quimby", "/"},
			{"admin", "/admin.html"},
			{u.Username, fmt.Sprintf("/admin/users/%s", u.Username)},
		}
		page.Actions = []action{
			{Name: "cancel", URI: template.URL("/admin.html"), Method: "get"},
			{Name: "delete", URI: template.URL(fmt.Sprintf("/admin/users/%s/delete", username)), Method: "get"},
			{Name: "update-password", URI: template.URL(fmt.Sprintf("/admin/users/%s/password", username)), Method: "get"},
			{Name: "update-tfa", URI: template.URL(fmt.Sprintf("/admin/users/%s/tfa", username)), Method: "post"},
		}
		page.End = 3
	}

	if username == "new-user" {
		templates["new-user.html"].template.ExecuteTemplate(w, "base", page)
	} else {
		templates["edit-user.html"].template.ExecuteTemplate(w, "base", page)
	}
}

func DeleteUserConfirmPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	u := quimby.NewUser(args.Vars["username"])
	page := editUserPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
			Links: []link{
				{"quimby", "/"},
				{"admin", "/admin.html"},
				{"new user", "/admin/users/new-user"},
			},
		},
		EditUser: u,
		Actions: []action{
			{Name: "cancel", URI: template.URL(fmt.Sprintf("/admin/users/%s", u.Username)), Method: "get"},
		},
	}
	templates["delete.html"].template.ExecuteTemplate(w, "base", page)
}

func DeleteUserPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	u := quimby.NewUser(args.Vars["username"], quimby.UserDB(args.DB))
	if err := u.Delete(); err != nil {
		context.Set(req, "error", err)
		return
	}
	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusMovedPermanently)
}

func UserPasswordPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	u := quimby.NewUser(args.Vars["username"])
	page := editUserPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
			Links: []link{
				{"quimby", "/"},
				{"admin", "/admin.html"},
				{u.Username, fmt.Sprintf("/admin/users/%s", u.Username)},
			},
		},
		EditUser: u,
		Actions: []action{
			{Name: "cancel", URI: template.URL(fmt.Sprintf("/admin/users/%s", u.Username)), Method: "get"},
		},
	}
	templates["password.html"].template.ExecuteTemplate(w, "base", page)
}

func UserChangePasswordPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	u := quimby.NewUser(args.Vars["username"], quimby.UserDB(args.DB))
	if err := u.Fetch(); err != nil {
		context.Set(req, "error", err)
		return
	}
	if err := req.ParseForm(); err != nil {
		context.Set(req, "error", err)
		return
	}

	u.Password = req.PostFormValue("password")
	pw := req.PostFormValue("password_confirm")
	if pw != u.Password {
		context.Set(req, "error", ErrPasswordsDoNotMatch)
		return
	}

	if _, err := u.Save(); err != nil {
		context.Set(req, "error", ErrPasswordsDoNotMatch)
		return
	}

	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusMovedPermanently)
}

func UserTFAPage(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)

	u := quimby.NewUser(args.Vars["username"], quimby.UserDB(args.DB), quimby.UserTFA(TFA))
	if err := u.Fetch(); err != nil {
		context.Set(req, "error", err)
		return
	}

	qrData, err := u.UpdateTFA()
	if err != nil {
		context.Set(req, "error", err)
		return
	}

	if _, err := u.Save(); err != nil {
		context.Set(req, "error", err)
		return
	}

	qr := qrPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
			Links: []link{
				{"quimby", "/"},
				{"admin", "/admin.html"},
			},
		},
		QR: template.HTMLAttr(base64.StdEncoding.EncodeToString(qrData)),
	}
	templates["qr-code.html"].template.ExecuteTemplate(w, "base", qr)
}

type qrPage struct {
	userPage
	QR template.HTMLAttr
}

func UserForm(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)

	err := req.ParseForm()
	if err != nil {
		context.Set(req, "error", err)
		return
	}

	u := quimby.NewUser(req.PostFormValue("username"), quimby.UserDB(args.DB), quimby.UserTFA(TFA))
	u.Password = req.PostFormValue("password")
	pw := req.PostFormValue("password_confirm")
	if pw != u.Password {
		context.Set(req, "error", ErrPasswordsDoNotMatch)
		return
	}
	u.Permission = req.PostFormValue("permission")
	qrData, err := u.Save()
	if err != nil {
		context.Set(req, "error", err)
		return
	}
	qr := qrPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: Admin(args),
			Links: []link{
				{"quimby", "/"},
				{"admin", "/admin.html"},
			},
		},
		QR: template.HTMLAttr(base64.StdEncoding.EncodeToString(qrData)),
	}
	templates["qr-code.html"].template.ExecuteTemplate(w, "base", qr)
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

	p := os.Getenv("QUIMBY_PORT")
	d := os.Getenv("QUIMBY_DOMAIN")
	var u string
	if p == "443" {
		u = fmt.Sprintf("wss://%s:%s", d, p)
	} else {
		u = fmt.Sprintf("ws://%s:%s", d, p)
	}

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
	templates["gadget.html"].template.ExecuteTemplate(w, "base", g)
}

func LoginPage(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	var p gadgetPage
	if q.Get("error") != "" {
		p.Error = "Invalid username or password"
	}
	templates["login.html"].template.ExecuteTemplate(w, "base", p)
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
	templates["logout"].template.ExecuteTemplate(w, "base", p)
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
