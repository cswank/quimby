package user

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"syscall"

	"github.com/cswank/quimby/internal/auth"
	"github.com/cswank/quimby/internal/repository"
	"golang.org/x/crypto/ssh/terminal"
)

func Create(r *repository.User, username string) error {
	fmt.Print("Enter Password: ")
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	pw, tfa, qa, err := auth.Credentials(username, string(pw))
	if err != nil {
		return err
	}

	u, err := r.Create(username, pw, tfa)
	if err != nil {
		return err
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(qa)
	}

	ts := httptest.NewServer(http.HandlerFunc(h))

	fmt.Printf("\ncreated user: %d %s\n, open %s to scan code\n", u.ID, u.Name, ts.URL)
	fmt.Println("hit enter when done")

	var i int
	fmt.Scanln(&i)

	ts.Close()

	return nil
}
