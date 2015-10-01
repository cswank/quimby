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
	LG.Println(r.URL.Path)
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
	if len(a.Vars["ticket"]) > 0 {
		a.doGetUserFromTicket()
	} else {
		a.doGetUserFromToken()
	}
	if a.err == nil {
		a.User.DB = a.DB
		a.err = a.User.Fetch()
		a.User.HashedPassword = []byte{}
	}
}

func (a *Args) doGetUserFromTicket() {
	var t ticket
	t, a.err = checkTicket(a.Vars["ticket"], a.R.Host)
	if a.err != nil {
		a.msg = "Not Authorized"
		a.status = http.StatusUnauthorized
	} else {
		a.User = &models.User{Username: t.user}
		a.Vars["id"] = t.id
	}
}

func (a *Args) doGetUserFromToken() {
	a.User, a.err = getUserFromToken(a.R)
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
	if a.err == nil {
		return
	}
	a.W.WriteHeader(a.status)
	a.W.Write([]byte(a.msg))
}