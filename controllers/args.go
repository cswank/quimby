package controllers

import (
	"errors"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/mux"
)

type Controller func(args *Args) error
type caller func()

type Args struct {
	W      http.ResponseWriter
	R      *http.Request
	DB     *bolt.DB
	User   *models.User
	Gadget *models.Gadget
	Vars   map[string]string
	LG     models.Logger
	acl    ACL
	ctrl   Controller
	err    error
	status int
	msg    string
}

func Handle(w http.ResponseWriter, r *http.Request, ctrl Controller, acl ACL) {
	LG.Println(r.URL.Path)
	a := &Args{
		W:    w,
		R:    r,
		Vars: mux.Vars(r),
		DB:   DB,
		acl:  acl,
		ctrl: ctrl,
	}

	for _, f := range a.calls() {
		if a.err == nil {
			f()
		}
	}
	a.finish()
}

type userGetter func(*http.Request) (*models.User, error)

func (a *Args) calls() []caller {
	return []caller{a.getUser, a.checkACL, a.getGadget, a.callCtrl}
}

func (a *Args) getUser() {
	f := getUserFromCookie

	if len(a.R.Header.Get("Authorization")) > 0 {
		f = getUserFromToken
	}
	a.User, a.err = f(a.R)
	if a.err != nil {
		a.msg = "Not Authorized"
		a.status = http.StatusUnauthorized
		return
	}
	a.User.DB = a.DB
	a.err = a.User.Fetch()
	a.User.HashedPassword = []byte{}
}

func (a *Args) checkACL() {
	if !a.acl(a) {
		a.err = errors.New("Not Authorized")
		a.msg = "Not Authorized"
		a.status = http.StatusUnauthorized
	}
}

func (a *Args) getGadget() {
	if a.Vars["id"] == "" {
		return
	}
	a.Gadget = &models.Gadget{
		DB: DB,
		Id: a.Vars["id"],
	}

	a.err = a.Gadget.Fetch()
	if a.err != nil {
		if a.err == models.NotFound {
			a.msg = "Not Found"
			a.status = http.StatusNotFound
		} else {
			a.status = http.StatusInternalServerError
			a.msg = "Internal Server Error"
		}
	}
}

func (a *Args) callCtrl() {
	a.err = a.ctrl(a)
	if a.err != nil {
		a.msg = a.err.Error()
		a.status = http.StatusInternalServerError
	}
}

func (a *Args) finish() {
	if a.err != nil {
		a.W.WriteHeader(a.status)
		a.W.Write([]byte(a.msg))
	}
}
