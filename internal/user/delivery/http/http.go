package userhttp

import (
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/cswank/quimby/internal/middleware"
	"github.com/cswank/quimby/internal/templates"
	"github.com/cswank/quimby/internal/user"
	"github.com/cswank/quimby/internal/user/usecase"
	"github.com/go-chi/chi"
)

// userHTTP renders html
type userHTTP struct {
	usecase user.Usecase
	box     *rice.Box
}

func Init(r chi.Router, box *rice.Box) {
	u := &userHTTP{
		usecase: usecase.New(),
		box:     box,
	}

	r.Get("/login", middleware.Handle(middleware.Render(u.renderLogin)))
	r.Post("/login", middleware.Handle(u.login))
}

func (u *userHTTP) renderLogin(w http.ResponseWriter, req *http.Request) (middleware.Renderer, error) {
	p := templates.NewPage("login", "login.ghtml")
	return &p, nil
}

func (u *userHTTP) login(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	username := req.Form.Get("username")
	pw := req.Form.Get("password")
	token := req.Form.Get("token")
	if err := u.usecase.Check(username, pw, token); err != nil {
		return err
	}

	cookie, err := middleware.GenerateCookie(username)
	if err != nil {
		return err
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, req, "/gadgets", http.StatusSeeOther)
	return nil
}

// func Logout(w http.ResponseWriter, r *http.Request) {
// 	cookie := &http.Cookie{
// 		Name:   "quimby",
// 		Value:  "",
// 		Path:   "/",
// 		MaxAge: -1,
// 	}
// 	http.SetCookie(w, cookie)
// 	args := GetArgs(r)
// 	if args.Args.Get("web") == "true" {
// 		w.Header().Set("Location", "/login.html")
// 		w.WriteHeader(http.StatusTemporaryRedirect)
// 	}
// }

// func Login(w http.ResponseWriter, r *http.Request) {
// 	user := quimby.NewUser("", quimby.UserTFA(TFA))
// 	dec := json.NewDecoder(r.Body)
// 	err := dec.Decode(user)
// 	if err != nil {
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 		return
// 	}
// 	if err := DoLogin(user, w, r); err != nil {
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 	}
// }

// func DoLogin(user *quimby.User, w http.ResponseWriter, req *http.Request) error {
// 	goodPassword, err := user.CheckPassword()
// 	if !goodPassword || err != nil {
// 		return fmt.Errorf("bad request")
// 	}

// 	params := req.URL.Query()
// 	methods, ok := params["auth"]
// 	user.TFAData = []byte{}
// 	if ok && methods[0] == "jwt" {
// 		setToken(w, user)
// 	} else {
// 		setCookie(w, user)
// 	}
// 	return nil
// }

// func setToken(w http.ResponseWriter, user *quimby.User) {
// 	token, err := quimby.GenerateToken(user.Username, exp)
// 	if err != nil {
// 		w.WriteHeader(http.StatusUnauthorized)
// 	} else {

// 		w.Header().Set("Authorization", token)
// 	}
// }

// func setCookie(w http.ResponseWriter, user *quimby.User) {

// }
