package handlers

import (
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby"
	"github.com/cswank/rex"
	"github.com/gorilla/context"
	"github.com/justinas/alice"
)

func GetArgs(r *http.Request) *Args {
	if args := context.Get(r, "args"); args != nil {
		return args.(*Args)
	}
	return nil
}

func setArgs(r *http.Request, args *Args) {
	context.Set(r, "args", args)
}

func Auth(db *bolt.DB, lg quimby.Logger, router *rex.Router, name string) alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			if req.URL.Path == "/api/login" || req.URL.Path == "/api/logout" {
				h.ServeHTTP(w, req)
				return
			}

			f := quimby.GetUserFromCookie

			if len(req.Header.Get("Authorization")) > 0 {
				f = quimby.GetUserFromToken
			}

			user, err := f(req)

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Not Authorized"))
				return
			}

			user.DB = db
			if err := user.Fetch(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal server error"))
				return
			}
			user.HashedPassword = []byte{}
			args := &Args{
				User: user,
				DB:   db,
				LG:   lg,
				Vars: rex.Vars(req, name),
				Args: req.URL.Query(),
			}

			setArgs(req, args)

			h.ServeHTTP(w, req)
		})
	}
}

func FetchGadget() alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			if req.URL.Path == "/api/login" || req.URL.Path == "/api/logout" {
				h.ServeHTTP(w, req)
				return
			}

			args := GetArgs(req)
			if args.Vars["id"] == "" {
				h.ServeHTTP(w, req)
				return
			}
			args.Gadget = &quimby.Gadget{
				DB: DB,
				Id: args.Vars["id"],
			}

			if err := args.Gadget.Fetch(); err != nil {
				var status int
				var msg string
				if err == quimby.NotFound {
					msg = "Not Found"
					status = http.StatusNotFound
				} else {
					status = http.StatusInternalServerError
					msg = "Internal Server Error"
				}
				w.WriteHeader(status)
				w.Write([]byte(msg))
				return
			}
			h.ServeHTTP(w, req)
		})
	}
}

type ACL func(*Args) bool

func Or(acls ...ACL) ACL {
	return func(args *Args) bool {
		for _, f := range acls {
			if f(args) {
				return true
			}
		}
		return false
	}
}

func And(acls ...ACL) ACL {
	return func(args *Args) bool {
		b := false
		for _, f := range acls {
			b = b && f(args)
		}
		return b
	}
}

func Admin(args *Args) bool {
	return args.User.Permission == "admin"
}

func Write(args *Args) bool {
	return args.User.Permission == "admin" || args.User.Permission == "write"
}

func Read(args *Args) bool {
	return args.User.Permission == "admin" || args.User.Permission == "write" || args.User.Permission == "read"
}

func Anyone(args *Args) bool {
	return true
}

func Perm(f ACL) alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			args := GetArgs(req)
			if !f(args) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Not Authorized"))
				return
			}
			h.ServeHTTP(w, req)
		})
	}
}
