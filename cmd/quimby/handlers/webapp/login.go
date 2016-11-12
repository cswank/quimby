package webapp

import (
	"html/template"
	"net/http"

	"github.com/cswank/quimby"
	"github.com/cswank/quimby/cmd/quimby/handlers"
)

func LoginPage(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	var p gadgetPage
	if q.Get("error") != "" {
		p.Error = "Invalid username or password"
	}
	templates["login.html"].template.ExecuteTemplate(w, "base", p)
}

func LoginForm(w http.ResponseWriter, req *http.Request) {
	user := quimby.NewUser("", quimby.UserDB(handlers.DB), quimby.UserTFA(handlers.TFA))
	user.Username = req.PostFormValue("username")
	user.Password = req.PostFormValue("password")
	user.TFA = req.PostFormValue("tfa")
	if err := handlers.DoLogin(user, w, req); err != nil {
		w.Header().Set("Location", "/login.html?error=invalidlogin")
	} else {
		w.Header().Set("Location", "/index.html")
	}
	w.WriteHeader(http.StatusFound)
}

func LogoutPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	p := editUserPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/"},
			},
		},
		Actions: []action{
			{Name: "cancel", URI: template.URL("/"), Method: "get"},
		},
	}
	templates["logout.html"].template.ExecuteTemplate(w, "base", p)
}
