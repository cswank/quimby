package admin

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
	"github.com/howeyc/gopass"
)

func ListUsers(db *bolt.DB) {
	users, err := models.GetUsers(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("# username  permission")
	for i, u := range users {
		fmt.Println(i+1, u.Username, u.Permission)
	}
}

func AddUser(db *bolt.DB) {
	u := models.User{
		DB: db,
	}
	fmt.Print("username: ")
	fmt.Scanf("%s", &u.Username)
	fmt.Print("can write? (y/N): ")
	var perm string
	fmt.Scanf("%s", &perm)
	if perm == "y" || perm == "Y" {
		u.Permission = "write"
	}
	fmt.Printf("password: ")
	p1 := string(gopass.GetPasswd())
	fmt.Printf("again: ")
	p2 := string(gopass.GetPasswd())
	if p1 != p2 {
		log.Fatal("passwords don't match")
	}
	u.Password = p1
	log.Println(u.Save())
}

func AddGadget(db *bolt.DB) {
	g := models.Gadget{
		DB: db,
	}
	fmt.Print("name: ")
	fmt.Scanf("%s", &g.Name)
	fmt.Print("host: ")
	fmt.Scanf("%s", &g.Host)
	fmt.Print(fmt.Sprintf("really save gadget (name: %s, host: %s)? (Y/n) ", g.Name, g.Host))
	var save string
	fmt.Scanf("%s", &save)
	if save == "y" || save == "Y" || save == "" {
		fmt.Println(g.Save())
	} else {
		fmt.Println("not saving")
	}
}

func ListGadgets(db *bolt.DB) {
	gadgets, err := models.GetGadgets(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("# name host")
	for i, g := range gadgets {
		fmt.Println(i+1, g.Name, g.Host)
	}
}
