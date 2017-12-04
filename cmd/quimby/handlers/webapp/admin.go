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

func AdminPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	gadgets, err := quimby.GetGadgets()
	if err != nil {
		return err
	}

	links := make([]link, len(gadgets))
	for i, g := range gadgets {
		links[i] = link{Name: g.Name, Path: fmt.Sprintf("/admin/gadgets/%s", g.Id)}
	}

	users, err := quimby.GetUsers()
	if err != nil {
		return err
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
				{"quimby", "/home"},
				{"admin", "/admin.html"},
			},
		},
		Gadgets: links,
		Users:   userLinks,
	}
	return templates["admin.html"].template.ExecuteTemplate(w, "base", i)
}

func GadgetEditPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	id := args.Vars["gadgetid"]
	g := &quimby.Gadget{Id: id}
	if id == "new-gadget" {
		g.Id = "new-gadget"
		g.Name = "new-gadget"
	} else {
		if err := g.Fetch(); err != nil {
			return err
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
				{"quimby", "/home"},
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
	return templates["edit-gadget.html"].template.ExecuteTemplate(w, "base", p)
}

func UserEditPage(w http.ResponseWriter, req *http.Request) error {
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
			{"quimby", "/home"},
			{"admin", "/admin.html"},
			{"new user", "/admin/users/new-user"},
		}
		page.Actions = []action{
			{Name: "cancel", URI: template.URL("/admin.html"), Method: "get"},
		}
		page.End = 0
	} else {
		u := quimby.NewUser(username)
		if err := u.Fetch(); err != nil {
			return err
		}

		q := url.Values{}
		q.Add("resource", fmt.Sprintf("/admin/users/%s", username))
		q.Add("name", username)
		page.EditUser = u
		page.Links = []link{
			{"quimby", "/home"},
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
		return templates["new-user.html"].template.ExecuteTemplate(w, "base", page)
	}
	return templates["edit-user.html"].template.ExecuteTemplate(w, "base", page)
}

func DeleteConfirmPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	page := confirmPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/home"},
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
	return templates["delete.html"].template.ExecuteTemplate(w, "base", page)
}

func DeleteUserPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	u := quimby.NewUser(args.Vars["username"])
	if err := u.Delete(); err != nil {
		return err
	}
	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusFound)
	return nil
}

func DeleteGadgetPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	g := &quimby.Gadget{Id: args.Vars["gadgetid"]}
	if err := g.Delete(); err != nil {
		return err
	}
	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusFound)
	return nil
}

func UserPasswordPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	u := quimby.NewUser(args.Vars["username"])
	page := editUserPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/home"},
				{"admin", "/admin.html"},
				{u.Username, fmt.Sprintf("/admin/users/%s", u.Username)},
			},
		},
		EditUser: u,
		Actions: []action{
			{Name: "cancel", URI: template.URL(fmt.Sprintf("/admin/users/%s", u.Username)), Method: "get"},
		},
	}
	return templates["password.html"].template.ExecuteTemplate(w, "base", page)
}

func UserChangePasswordPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	u := quimby.NewUser(args.Vars["username"])
	if err := u.Fetch(); err != nil {
		return err
	}
	if err := req.ParseForm(); err != nil {
		return err
	}

	u.Password = req.PostFormValue("password")
	pw := req.PostFormValue("password_confirm")
	if pw != u.Password {
		return fmt.Errorf("invalid password")
	}

	if _, err := u.Save(); err != nil {
		return err
	}

	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusFound)
	return nil
}

func UserTFAPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	u := quimby.NewUser(args.Vars["username"], quimby.UserTFA(handlers.TFA))
	if err := u.Fetch(); err != nil {
		return err
	}

	qrData, err := u.UpdateTFA()
	if err != nil {
		return err
	}

	if _, err := u.Save(); err != nil {
		return err
	}

	qr := qrPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/home"},
				{"admin", "/admin.html"},
			},
		},
		QR: template.HTMLAttr(base64.StdEncoding.EncodeToString(qrData)),
	}
	return templates["qr-code.html"].template.ExecuteTemplate(w, "base", qr)
}

type qrPage struct {
	userPage
	QR template.HTMLAttr
}

func UserForm(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)

	err := req.ParseForm()
	if err != nil {
		return err
	}

	username := args.Vars["username"]
	var u *quimby.User
	if username == "new-user" {
		u = quimby.NewUser(req.PostFormValue("username"), quimby.UserTFA(handlers.TFA))
		u.Password = req.PostFormValue("password")
		pw := req.PostFormValue("password_confirm")
		if pw != u.Password {
			return err
		}
	} else {
		u = quimby.NewUser(username, quimby.UserTFA(handlers.TFA))
		if err := u.Fetch(); err != nil {
			return err
		}
	}
	u.Permission = req.PostFormValue("permission")
	qrData, err := u.Save()
	if err != nil {
		return err
	}
	if username == "new-user" {
		qr := qrPage{
			userPage: userPage{
				User:  args.User.Username,
				Admin: handlers.Admin(args),
				Links: []link{
					{"quimby", "/home"},
					{"admin", "/admin.html"},
				},
			},
			QR: template.HTMLAttr(base64.StdEncoding.EncodeToString(qrData)),
		}
		return templates["qr-code.html"].template.ExecuteTemplate(w, "base", qr)
	}

	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusFound)
	return nil
}

func GadgetForm(w http.ResponseWriter, req *http.Request) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}
	args := handlers.GetArgs(req)
	g := &quimby.Gadget{}
	if args.Vars["gadgetid"] != "new-gadget" {
		g.Id = args.Vars["gadgetid"]
	}

	g.Host = req.PostFormValue("host")
	g.Name = req.PostFormValue("name")
	g.View = req.PostFormValue("view")
	s := req.PostFormValue("disabled")

	d := s == "on"
	g.Disabled = d
	context.Set(req, "error", g.Save())
	w.Header().Set("Location", "/admin.html")
	w.WriteHeader(http.StatusFound)
	return nil
}
