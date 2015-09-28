package controllers

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

func Write(args *Args) bool {
	return args.User.Permission == "write" || args.User.Permission == "admin"
}

func Read(args *Args) bool {
	return args.User.Permission == "write" || args.User.Permission == "admin" || args.User.Permission == "read"
}

func Anyone(args *Args) bool {
	return true
}
