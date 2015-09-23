package models

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username       string   `json:"username"`
	Password       string   `json:"password,omitempty"`
	HashedPassword []byte   `json:"hashed_password,omitempty"`
	Permission     string   `json:"permission"`
	DB             *bolt.DB `json:"-"`
}

func GetUsers(db *bolt.DB) ([]User, error) {
	users := []User{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var u User
			if err := json.Unmarshal(v, &u); err != nil {
				return err
			}
			u.HashedPassword = []byte{}
			users = append(users, u)
		}
		return nil
	})
	return users, err
}

func (u *User) Fetch() error {
	return u.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		v := b.Get([]byte(u.Username))
		err := json.Unmarshal(v, u)
		return err
	})
}

//Is authorized if the username is in the db
func (u *User) IsAuthorized(perm string) bool {
	if u.Permission == "" {
		if err := u.Fetch(); err != nil {
			return false
		}
	}
	return u.Permission == perm && perm != ""
}

func (u *User) Save() error {
	if len(u.Password) < 8 {
		return errors.New("password is too short")
	}
	u.hashPassword()
	return u.DB.Update(func(tx *bolt.Tx) error {
		d, _ := json.Marshal(u)
		b := tx.Bucket([]byte("users"))
		return b.Put([]byte(u.Username), d)
	})
}

func (u *User) Delete() error {
	return u.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		return b.Delete([]byte(u.Username))
	})
}

func (u *User) CheckPassword() (bool, error) {
	pw := u.Password
	if len(u.HashedPassword) == 0 {
		if err := u.Fetch(); err != nil {
			return false, err
		}
	}
	return bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(pw)) == nil, nil
}

func (u *User) hashPassword() {
	u.HashedPassword, _ = bcrypt.GenerateFromPassword(
		[]byte(u.Password),
		bcrypt.DefaultCost,
	)
	u.Password = ""
}
