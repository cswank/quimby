package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/controllers"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

type controller func(args *controllers.Args) error

var (
	DB           *bolt.DB
	hashKey      = []byte(os.Getenv("QUIMBY_HASH_KEY"))
	blockKey     = []byte(os.Getenv("QUIMBY_BLOCK_KEY"))
	SecureCookie = securecookie.New(hashKey, blockKey)
)

func CheckAuth(w http.ResponseWriter, r *http.Request, ctrl controller, acl ACL) {
	user, err := getUserFromCookie(r)
	if err != nil {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	args := &controllers.Args{
		W:    w,
		R:    r,
		User: user,
		Vars: vars,
		DB:   DB,
	}
	if !acl(args) {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		return
	}

	g := &models.Gadget{
		DB:   DB,
		Name: vars["name"],
	}

	args.Gadget = g
	err = ctrl(args)
	if err != nil {
		log.Println(args.R.URL.Path, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getUserFromCookie(r *http.Request) (*models.User, error) {
	user := &models.User{
		DB: DB,
	}
	cookie, err := r.Cookie("quimby")
	if err != nil {
		return nil, err
	}
	var m map[string]string
	err = SecureCookie.Decode("quimby", cookie.Value, &m)
	if err != nil {
		return nil, err
	}
	user.Username = m["user"]
	return user, user.Fetch()
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "quimby",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

func Login(w http.ResponseWriter, r *http.Request) {
	user := &models.User{
		DB: DB,
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(user)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	goodPassword, err := user.CheckPassword()
	fmt.Println("good?", goodPassword, err)
	if !goodPassword {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	value := map[string]string{
		"user": user.Username,
	}

	encoded, _ := SecureCookie.Encode("quimby", value)
	cookie := &http.Cookie{
		Name:     "quimby",
		Value:    encoded,
		Path:     "/",
		HttpOnly: false,
	}
	http.SetCookie(w, cookie)
}
