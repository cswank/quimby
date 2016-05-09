package utils

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby"
	"github.com/howeyc/gopass"
)

func listUsers(users []quimby.User) {
	fmt.Println("# username  permission")
	for i, u := range users {
		fmt.Println(i+1, u.Username, u.Permission)
	}
}

func ListUsers(db *bolt.DB) {
	users, err := quimby.GetUsers(db)
	if err != nil {
		log.Fatal(err)
	}
	listUsers(users)
}

func EditUser(db *bolt.DB) {
	users, err := quimby.GetUsers(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("select a user")
	listUsers(users)

	var i int
	fmt.Scanf("%d\n", &i)
	u := users[i-1]
	u.DB = db

	var p int
	fmt.Printf("permission (%s):\n  1: read\n  2: write\n  3: admin\n ", u.Permission)
	fmt.Scanf("%d\n", &p)
	perm, ok := permissions[p]
	if !ok {
		log.Fatal("select 1, 2, or 3")
	}
	u.Permission = perm

	var c string
	fmt.Print("change password? (y/N) ")
	fmt.Scanf("%s\n", &c)
	if c == "y" || c == "Y" {
		getPasswd(&u)
	}
	log.Println(u.Save())
}

var (
	permissions = map[int]string{
		1: "read",
		2: "write",
		3: "admin",
		4: "sys",
	}
)

func getPasswd(u *quimby.User) {
	fmt.Printf("password: ")
	p1 := string(gopass.GetPasswd())
	fmt.Printf("again: ")
	p2 := string(gopass.GetPasswd())
	if p1 != p2 {
		log.Fatal("passwords don't match")
	}
	u.Password = p1
}

type passworder func(*quimby.User)

func genPasswd(u *quimby.User) {
	u.Password = randString(32)
}

func AddUser(u *quimby.User) {
	var f passworder
	if u.Username == "" {
		fmt.Print("username: ")
		fmt.Scanf("%s\n", &u.Username)
		fmt.Print("permission:\n  1: read\n  2: write\n  3: admin\n  4: system\n")
		var x int
		fmt.Scanf("%d\n", &x)
		if x == 4 {
			f = genPasswd
		} else {
			f = getPasswd
		}
		perm, ok := permissions[x]
		if !ok {
			log.Fatal("select 1, 2, 3, or 4")
		}
		u.Permission = perm
		f(u)
	}
	log.Println(u.Save())
}

func AddGadget(g *quimby.Gadget) {
	if g.Name == "" {
		fmt.Print("name: ")
		fmt.Scanf("%s\n", &g.Name)
		fmt.Print("host: ")
		fmt.Scanf("%s\n", &g.Host)
		fmt.Print(fmt.Sprintf("really save gadget (name: %s, host: %s)? (Y/n) ", g.Name, g.Host))
		var save string
		fmt.Scanf("%s\n", &save)
		if save == "y" || save == "Y" || save == "" {
			fmt.Println(g.Save())
		} else {
			fmt.Println("not saving")
		}
	} else {
		err := g.Save()
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(g.Id)
}

func DeleteGadget(db *bolt.DB) {
	gadgets, err := quimby.GetGadgets(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("pick a number")
	listGadgets(gadgets)
	var n int
	fmt.Scanf("%d\n", &n)
	g := gadgets[n-1]
	g.DB = db
	fmt.Println(g.Delete())
}

func EditGadget(db *bolt.DB) {
	gadgets, err := quimby.GetGadgets(db)
	if err != nil {
		log.Fatal(err)
	}
	listGadgets(gadgets)
	var i int
	fmt.Scanf("%d\n", &i)
	g := gadgets[i-1]
	g.DB = db

	var n string
	fmt.Printf("name (%s): ", g.Name)
	fmt.Scanf("%s\n", &n)
	fmt.Printf("host (%s): ", g.Host)
	fmt.Scanf("%s\n", &g.Host)
	if n != g.Name && n != "" {
		g.Delete()
		g.Name = n
	}
	fmt.Println(g.Save())
}

func ListGadgets(db *bolt.DB) {
	gadgets, err := quimby.GetGadgets(db)
	if err != nil {
		log.Fatal(err)
	}
	listGadgets(gadgets)
}

func listGadgets(gadgets []quimby.Gadget) {
	fmt.Println("# name host")
	for i, g := range gadgets {
		fmt.Println(i+1, g.Name, g.Host, g.Id)
	}
}
