package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/cswank/quimby"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func GetArgs(r *http.Request) *Args {
	return r.Context().Value("args").(*Args)
}

func setArgs(r *http.Request, args *Args) *http.Request {
	ctx := context.WithValue(r.Context(), "args", args)
	return r.WithContext(ctx)
}

func Error(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := h(w, req)
		if err != nil {
			log.Printf("error (%s)\n", err)
			if err.Error() == "not found" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write([]byte(err.Error()))
		}
	}
}

func Auth() alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			pth := req.URL.Path
			if pth == "/api/login" || strings.Index(pth, "/css") == 0 || strings.Index(pth, "/login.html") == 0 {
				h.ServeHTTP(w, req)
				return
			}

			f := quimby.GetUserFromCookie
			if len(req.Header.Get("Authorization")) > 0 {
				f = quimby.GetUserFromToken
			}

			user, err := f(req)
			if err != nil {
				if strings.Index(pth, "/api") > -1 {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("Not Authorized"))
				} else {
					w.Header().Set("Location", "/login.html")
					w.WriteHeader(http.StatusMovedPermanently)
				}
				return
			}

			if err := user.Fetch(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal server error"))
				return
			}
			user.HashedPassword = []byte{}
			args := &Args{
				User: user,
				Vars: mux.Vars(req),
				Args: req.URL.Query(),
			}

			req = setArgs(req, args)
			h.ServeHTTP(w, req)
		})
	}
}

func FetchGadget() alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			pth := req.URL.Path
			if pth == "/api/login" || strings.Index(pth, "/css") == 0 || strings.Index(pth, "/login.html") == 0 {
				h.ServeHTTP(w, req)
				return
			}

			args := GetArgs(req)
			if args == nil || args.Vars["id"] == "" {
				h.ServeHTTP(w, req)
				return
			}
			args.Gadget = &quimby.Gadget{
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
