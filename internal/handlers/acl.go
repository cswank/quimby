package handlers

import "log"

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

func Twillo(args *Args) bool {
	log.Println(args.R.Header)
	return true
}

func Read(args *Args) bool {
	return args.User.Permission == "admin" || args.User.Permission == "write" || args.User.Permission == "read"
}

func Anyone(args *Args) bool {
	return true
}
