package user

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"syscall"

	"github.com/cswank/quimby/internal/auth"
	"github.com/cswank/quimby/internal/repository"
	"golang.org/x/crypto/ssh/terminal"
)

func Create(r *repository.User, username string) {
	fmt.Print("Enter password: ")
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Confirm password: ")
	pw2, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}

	if !bytes.Equal(pw, pw2) {
		log.Fatal(fmt.Errorf("passwords don't match"))
	}

	hashed, tfa, qr, err := auth.Credentials(username, string(pw))
	if err != nil {
		log.Fatal(err)
	}

	u, err := r.Create(username, hashed, tfa)
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(qr)
	}))

	fmt.Printf("\ncreated user: %d %s\n, open %s to scan code\n", u.ID, u.Name, ts.URL)
	fmt.Println("hit enter when done")

	var i int
	fmt.Scanln(&i)

	ts.Close()
}

func Delete(r *repository.User, username string) {
	if err := r.Delete(username); err != nil {
		log.Fatal(err)
	}
}

func List(r *repository.User) {
	users, err := r.GetAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		fmt.Printf("%d: %s\n", u.ID, u.Name)
	}
}
