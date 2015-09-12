package auth

import "github.com/cswank/quimby/controllers"

type ACL func(*controllers.Args) bool

func Or(acls ...ACL) ACL {
	return func(args *controllers.Args) bool {
		for _, f := range acls {
			if f(args) {
				return true
			}
		}
		return false
	}
}

func And(acls ...ACL) ACL {
	return func(args *controllers.Args) bool {
		b := false
		for _, f := range acls {
			b = b && f(args)
		}
		return b
	}
}

func Write(args *controllers.Args) bool {
	return args.User.Permission == "write" || args.User.Permission == "admin"
}

func Read(args *controllers.Args) bool {
	return args.User.Permission == "write" || args.User.Permission == "admin" || args.User.Permission == "read"
}
