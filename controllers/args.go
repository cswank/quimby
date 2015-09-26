package controllers

import (
	"errors"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/mux"
)

type controller func(args *Args) error
type caller func()

type Args struct {
	W      http.ResponseWriter
	R      *http.Request
	DB     *bolt.DB
	User   *models.User
	Gadget *models.Gadget
	Vars   map[string]string
	LG     Logger
	acl    ACL
	ctrl   controller
	err    error
	status int
	msg    string
}

func Handle(w http.ResponseWriter, r *http.Request, ctrl controller, acl ACL) {
	a := &Args{
		W:    w,
		R:    r,
		Vars: mux.Vars(r),
		DB:   DB,
		acl:  acl,
		ctrl: ctrl,
	}
	calls := []caller{a.getUser, a.checkACL, a.getGadget, a.callCtrl}
	for _, f := range calls {
		if a.err == nil {
			f()
		}
	}
	a.finish()
}

func (a *Args) getUser() {
	a.User, a.err = getUserFromCookie(a.R)
	if a.err != nil {
		a.msg = "Not Authorized"
		a.status = http.StatusUnauthorized
	}
}

func (a *Args) checkACL() {
	if !a.acl(a) {
		a.err = errors.New("Not Authorized")
		a.msg = "Not Authorized"
		a.status = http.StatusUnauthorized
	}
}

func (a *Args) getGadget() {
	if a.Vars["name"] == "" {
		return
	}
	a.Gadget = &models.Gadget{
		DB:   DB,
		Name: a.Vars["name"],
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
	if a.err != nil {
		return
	}
	a.err = a.ctrl(a)
	if a.err != nil {
		a.msg = a.err.Error()
		a.status = http.StatusInternalServerError
	}
}

func (a *Args) finish() {
	if a.err == nil {
		return
	}
	a.W.WriteHeader(a.status)
	a.W.Write([]byte(a.msg))
}
