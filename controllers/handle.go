package controllers

import (
	"errors"
	"net/http"

	"github.com/cswank/quimby/models"
	"github.com/gorilla/mux"
)

type controller func(args *Args) error

func Handle(w http.ResponseWriter, r *http.Request, ctrl controller, acl ACL) {
	h := &handler{w: w, r: r, acl: acl, ctrl: ctrl}
	calls := []handlerer{h.getUser, h.getArgs, h.checkACL, h.getGadget, h.callCtrl}
	for _, f := range calls {
		if h.err == nil {
			f()
		}
	}
	h.finish()
}

type handlerer func()

type handler struct {
	w      http.ResponseWriter
	r      *http.Request
	acl    ACL
	ctrl   controller
	err    error
	status int
	msg    string
	user   *models.User
	args   *Args
}

func (h *handler) getUser() {
	h.user, h.err = getUserFromCookie(h.r)
	if h.err != nil {
		h.msg = "Not Authorized"
		h.status = http.StatusUnauthorized
	}
}

func (h *handler) getArgs() {
	h.args = &Args{
		W:    h.w,
		R:    h.r,
		User: h.user,
		Vars: mux.Vars(h.r),
		DB:   DB,
	}
}

func (h *handler) checkACL() {
	if !h.acl(h.args) {
		h.err = errors.New("Not Authorized")
		h.msg = "Not Authorized"
		h.status = http.StatusUnauthorized
	}
}

func (h *handler) getGadget() {
	if h.args.Vars["name"] == "" {
		return
	}
	h.args.Gadget = &models.Gadget{
		DB:   DB,
		Name: h.args.Vars["name"],
	}

	h.err = h.args.Gadget.Fetch()
	if h.err != nil {
		if h.err == models.NotFound {
			h.msg = "Not Found"
			h.status = http.StatusNotFound
		} else {
			h.status = http.StatusInternalServerError
			h.msg = "Internal Server Error"
		}
	}
}

func (h *handler) callCtrl() {
	if h.err != nil {
		return
	}
	h.err = h.ctrl(h.args)
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
