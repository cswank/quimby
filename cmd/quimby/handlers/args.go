package handlers

import (
	"net/url"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby"
)

type Controller func(args *Args) error
type caller func()

type Args struct {
	DB     *bolt.DB
	User   *quimby.User
	Gadget *quimby.Gadget
	Vars   map[string]string
	Args   url.Values
	LG     quimby.Logger
}

// func Handle(w http.ResponseWriter, r *http.Request, ctrl Controller, acl ACL, name string) {
// 	LG.Println(r.URL.Path)
// 	a := &Args{
// 		W:    w,
// 		R:    r,
// 		Vars: rex.Vars(r, name),
// 		Args: r.URL.Query(),
// 		DB:   DB,
// 		acl:  acl,
// 		ctrl: ctrl,
// 	}

// 	for _, f := range a.calls() {
// 		if a.err == nil {
// 			f()
// 		}
// 	}
// 	a.finish()
// }

// type userGetter func(*http.Request) (*quimby.User, error)

// func (a *Args) calls() []caller {
// 	return []caller{a.getUser, a.checkACL, a.getGadget, a.callCtrl}
// }

// func (a *Args) getUser() {
// 	f := quimby.GetUserFromCookie

// 	if len(a.R.Header.Get("Authorization")) > 0 {
// 		f = quimby.GetUserFromToken
// 	}
// 	a.User, a.err = f(a.R)
// 	if a.err != nil {
// 		a.msg = "Not Authorized"
// 		a.status = http.StatusUnauthorized
// 		return
// 	}
// 	a.User.DB = a.DB
// 	a.err = a.User.Fetch()
// 	a.User.HashedPassword = []byte{}
// }

// func (a *Args) checkACL() {
// 	if !a.acl(a) {
// 		a.err = errors.New("Not Authorized")
// 		a.msg = "Not Authorized"
// 		a.status = http.StatusUnauthorized
// 	}
// }

// func (a *Args) getGadget() {
// 	if a.Vars["id"] == "" {
// 		return
// 	}
// 	a.Gadget = &quimby.Gadget{
// 		DB: DB,
// 		Id: a.Vars["id"],
// 	}

// 	a.err = a.Gadget.Fetch()
// 	if a.err != nil {
// 		if a.err == quimby.NotFound {
// 			a.msg = "Not Found"
// 			a.status = http.StatusNotFound
// 		} else {
// 			a.status = http.StatusInternalServerError
// 			a.msg = "Internal Server Error"
// 		}
// 	}
// }

// func (a *Args) callCtrl() {
// 	a.err = a.ctrl(a)
// 	if a.err != nil {
// 		a.msg = a.err.Error()
// 		if a.msg == "bad request" {
// 			a.status = http.StatusBadRequest
// 		} else {
// 			a.status = http.StatusInternalServerError
// 		}
// 	}
// }

// func (a *Args) finish() {
// 	if a.err != nil {
// 		LG.Println("error", a.err)
// 		a.W.WriteHeader(a.status)
// 		a.W.Write([]byte(a.msg))
// 	}
// }
