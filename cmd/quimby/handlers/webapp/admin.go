package webapp

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/cswank/quimby"
	"github.com/cswank/quimby/cmd/quimby/handlers"
	"github.com/gorilla/context"
)

func AdminPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
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
			Admin: handlers.Admin(args),
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

func GadgetEditPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	id := args.Vars["gadgetid"]
	g := &quimby.Gadget{Id: id, DB: args.DB}
	if id == "new-gadget" {
		g.Id = "new-gadget"
		g.Name = "new-gadget"
	} else {
		if err := g.Fetch(); err != nil {
			context.Set(req, "error", err)
			return
		}
	}

	q := url.Values{}
	q.Add("resource", fmt.Sprintf("/admin/gadgets/%s", g.Id))
	q.Add("name", g.Name)

	p := gadgetPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/"},
				{"admin", "/admin.html"},
				{g.Name, fmt.Sprintf("/admin/gadgets/%s", g.Id)},
			},
		},
		Gadget: g,
		Actions: []action{
			{Name: "cancel", URI: template.URL("/admin.html"), Method: "get"},
			{Name: "delete", URI: template.URL(fmt.Sprintf("/admin/confirmation?%s", q.Encode())), Method: "get"},
		},
		End: 2,
	}
	templates["edit-gadget.html"].template.ExecuteTemplate(w, "base", p)
}

func UserEditPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	username := args.Vars["username"]
	page := editUserPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
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

		q := url.Values{}
		q.Add("resource", fmt.Sprintf("/admin/users/%s", username))
		q.Add("name", username)
		page.EditUser = u
		page.Links = []link{
			{"quimby", "/"},
			{"admin", "/admin.html"},
			{u.Username, fmt.Sprintf("/admin/users/%s", u.Username)},
		}
		page.Actions = []action{
			{Name: "cancel", URI: template.URL("/admin.html"), Method: "get"},
			{Name: "delete", URI: template.URL(fmt.Sprintf("/admin/confirmation?%s", q.Encode())), Method: "get"},
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

func DeleteConfirmPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	page := confirmPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/"},
				{"admin", "/admin.html"},
				{"new user", "/admin/users/new-user"},
			},
		},
		Actions: []action{
			{Name: "delete", URI: template.URL(args.Args.Get("resource")), Method: "delete"},
		},
		Resource: args.Args.Get("resource"),
		Name:     args.Args.Get("name"),
	}
	templates["delete.html"].template.ExecuteTemplate(w, "base", page)
}

func DeleteUserPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	u := quimby.NewUser(args.Vars["username"], quimby.UserDB(args.DB))
	if err := u.Delete(); err != nil {
		context.Set(req, "error", err)
		return
	}
	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusFound)
}

func DeleteGadgetPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	g := &quimby.Gadget{Id: args.Vars["gadgetid"], DB: args.DB}
	if err := g.Delete(); err != nil {
		context.Set(req, "error", err)
		return
	}
	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusFound)
}

func UserPasswordPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	u := quimby.NewUser(args.Vars["username"])
	page := editUserPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
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
	args := handlers.GetArgs(req)
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
	w.WriteHeader(http.StatusFound)
}

func UserTFAPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)

	u := quimby.NewUser(args.Vars["username"], quimby.UserDB(args.DB), quimby.UserTFA(handlers.TFA))
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
			Admin: handlers.Admin(args),
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
	args := handlers.GetArgs(req)

	err := req.ParseForm()
	if err != nil {
		context.Set(req, "error", err)
		return
	}

	username := args.Vars["username"]
	var u *quimby.User
	if username == "new-user" {
		u = quimby.NewUser(req.PostFormValue("username"), quimby.UserDB(args.DB), quimby.UserTFA(handlers.TFA))
		u.Password = req.PostFormValue("password")
		pw := req.PostFormValue("password_confirm")
		if pw != u.Password {
			context.Set(req, "error", ErrPasswordsDoNotMatch)
			return
		}
	} else {
		u = quimby.NewUser(username, quimby.UserDB(args.DB), quimby.UserTFA(handlers.TFA))
		if err := u.Fetch(); err != nil {
			context.Set(req, "error", ErrPasswordsDoNotMatch)
			return
		}
	}
	u.Permission = req.PostFormValue("permission")
	qrData, err := u.Save()
	if err != nil {
		context.Set(req, "error", err)
		return
	}
	if username == "new-user" {
		qr := qrPage{
			userPage: userPage{
				User:  args.User.Username,
				Admin: handlers.Admin(args),
				Links: []link{
					{"quimby", "/"},
					{"admin", "/admin.html"},
				},
			},
			QR: template.HTMLAttr(base64.StdEncoding.EncodeToString(qrData)),
		}
		templates["qr-code.html"].template.ExecuteTemplate(w, "base", qr)
	} else {
		w.Header().Set("Location", "/admin.html")
		w.WriteHeader(http.StatusFound)
	}
}

func GadgetForm(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		context.Set(req, "error", err)
		return
	}
	args := handlers.GetArgs(req)
	g := &quimby.Gadget{DB: args.DB}
	if args.Vars["gadgetid"] != "new-gadget" {
		g.Id = args.Vars["gadgetid"]
	}

	g.Host = req.PostFormValue("host")
	g.Name = req.PostFormValue("name")
	s := req.PostFormValue("disabled")

	d := s == "on"
	g.Disabled = d
	context.Set(req, "error", g.Save())
	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusFound)
}
