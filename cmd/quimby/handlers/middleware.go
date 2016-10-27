package handlers

import (
	"fmt"
	"net/http"
	"strings"

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

func Error(lg quimby.Logger) alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(w, req)
			e := context.Get(req, "error")
			if e != nil {
				lg.Printf("error (%s)\n", e)
				err := e.(error)
				if err.Error() == "not found" {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
				w.Write([]byte(err.Error()))
			}
		})
	}
}

func Auth(db *bolt.DB, lg quimby.Logger, name string) alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			pth := req.URL.Path
			if pth == "/api/login" || strings.Index(pth, "/css") == 0 {
				h.ServeHTTP(w, req)
				return
			}

			f := quimby.GetUserFromCookie
			if len(req.Header.Get("Authorization")) > 0 {
				f = quimby.GetUserFromToken
			}

			user, err := f(req)
			if err != nil {
				fmt.Println("xxx")
				w.Header().Set("Location", "/login.html")
				w.WriteHeader(http.StatusMovedPermanently)
				//w.Write([]byte("Not Authorized"))
				return
			}

			user.SetDB(db)

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

			context.Clear(req)
		})
	}
}

func FetchGadget() alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			pth := req.URL.Path
			if pth == "/api/login" || (strings.Index(pth, "/api") == -1 && strings.Index(pth, "/internal") == -1) {
				h.ServeHTTP(w, req)
				return
			}

			args := GetArgs(req)
			if args == nil || args.Vars["id"] == "" {
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
			} else {
				setArgs(req, args)
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
	return args != nil && args.User != nil && args.User.Permission == "admin"
}

func Write(args *Args) bool {
	return args != nil && args.User != nil && (args.User.Permission == "admin" || args.User.Permission == "write")
}

func Read(args *Args) bool {
	return args != nil && args.User != nil && (args.User.Permission == "admin" || args.User.Permission == "write" || args.User.Permission == "read")
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
