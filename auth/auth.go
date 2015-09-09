package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cswank/gadgetsweb/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

type controller func(w http.ResponseWriter, r *http.Request, u *models.User, vars map[string]string) error

var (
	hashKey      = []byte(os.Getenv("GADGETS_HASH_KEY"))
	blockKey     = []byte(os.Getenv("GADGETS_BLOCK_KEY"))
	SecureCookie = securecookie.New(hashKey, blockKey)
)

func CheckAuth(w http.ResponseWriter, r *http.Request, ctrl controller, permission string) {
	user, err := getUserFromCookie(r)
	if err == nil && user.IsAuthorized(permission) {
		vars := mux.Vars(r)
		err = ctrl(w, r, user, vars)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		log.Println(err)
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
	}
}

func getUserFromCookie(r *http.Request) (*models.User, error) {
	user := &models.User{}
	cookie, err := r.Cookie("gadgets")
	if err == nil {
		m := map[string]string{}
		err = SecureCookie.Decode("gadgets", cookie.Value, &m)
		if err == nil {
			user.Username = m["user"]
		}
	}
	return user, err
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "gadgets",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

func Login(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(user)
	if err != nil {
		http.Error(w, "bad request 1", http.StatusBadRequest)
		return
	}
	goodPassword, err := user.CheckPassword()
	if !goodPassword {
		http.Error(w, "bad request 2", http.StatusBadRequest)
		return
	}
	value := map[string]string{
		"user": user.Username,
	}

	encoded, err := SecureCookie.Encode("gadgets", value)
	cookie := &http.Cookie{
		Name:     "gadgets",
		Value:    encoded,
		Path:     "/",
		HttpOnly: false,
	}
	fmt.Println("cookie", cookie, encoded)
	http.SetCookie(w, cookie)
}
