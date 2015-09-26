package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/controllers"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/mux"
)

type controller func(args *controllers.Args) error

var (
	DB *bolt.DB
)

func CheckAuth(w http.ResponseWriter, r *http.Request, ctrl controller, acl ACL) {
	h := &handler{w: w, r: r, acl: acl, ctrl: ctrl}
	user := h.getUser()
	args := h.getArgs(user)
	h.checkACL(args)
	h.getGadget(args)
	h.callCtrl(args)
	h.finish()
}

type handler struct {
	w      http.ResponseWriter
	r      *http.Request
	acl    ACL
	ctrl   controller
	err    error
	status int
	msg    string
}

func (h *handler) getUser() *models.User {
	u, err := getUserFromCookie(h.r)
	if err != nil {
		h.err = err
		h.msg = "Not Authorized"
		h.status = http.StatusUnauthorized
	}
	return u
}

func (h *handler) getArgs(u *models.User) *controllers.Args {
	if h.err != nil {
		return nil
	}
	return &controllers.Args{
		W:    h.w,
		R:    h.r,
		User: u,
		Vars: mux.Vars(h.r),
		DB:   DB,
	}
}

func (h *handler) checkACL(args *controllers.Args) {
	if h.err != nil {
		return
	}
	if !h.acl(args) {
		h.err = errors.New("Not Authorized")
		h.msg = "Not Authorized"
		h.status = http.StatusUnauthorized
	}
}

func (h *handler) getGadget(args *controllers.Args) {
	if h.err != nil || args.Vars["name"] == "" {
		return
	}

	g := &models.Gadget{
		DB:   DB,
		Name: args.Vars["name"],
	}

	h.err = g.Fetch()
	if h.err != nil {
		if h.err == models.NotFound {
			h.msg = "Not Found"
			h.status = http.StatusNotFound
		} else {
			h.status = http.StatusInternalServerError
			h.msg = "Internal Server Error"
		}
	}
	args.Gadget = g
}

func (h *handler) callCtrl(args *controllers.Args) {
	if h.err != nil {
		return
	}
	h.err = h.ctrl(args)
	if h.err != nil {
		h.msg = h.err.Error()
		h.status = http.StatusInternalServerError
	}
}

func (h *handler) finish() {
	if h.err == nil {
		return
	}
	h.w.WriteHeader(h.status)
	h.w.Write([]byte(h.msg))
}

func getUserFromCookie(r *http.Request) (*models.User, error) {
	user := &models.User{
		DB: DB,
	}
	cookie, err := r.Cookie("quimby")
	if err != nil {
		return nil, err
	}
	var m map[string]string
	err = controllers.SecureCookie.Decode("quimby", cookie.Value, &m)
	if err != nil {
		return nil, err
	}
	user.Username = m["user"]
	err = user.Fetch()
	user.HashedPassword = []byte{}
	return user, err
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "quimby",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

func Login(w http.ResponseWriter, r *http.Request) {
	user := &models.User{
		DB: DB,
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(user)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	goodPassword, err := user.CheckPassword()
	if !goodPassword {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	value := map[string]string{
		"user": user.Username,
	}

	encoded, _ := controllers.SecureCookie.Encode("quimby", value)
	cookie := &http.Cookie{
		Name:     "quimby",
		Value:    encoded,
		Path:     "/",
		HttpOnly: false,
	}
	w.Header().Set("Location", "/api/users/current")
	http.SetCookie(w, cookie)
}
