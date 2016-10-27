package quimby

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/boltdb/bolt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username       string `json:"username"`
	Password       string `json:"password,omitempty"`
	HashedPassword []byte `json:"hashed_password,omitempty"`
	TFA            string `json:"tfa,omitempty"`
	TFAData        []byte `json:"tfa_data,omitempty"`
	Permission     string `json:"permission"`
	db             *bolt.DB
	tfa            TFAer
}

type UserOpt func(*User)

func UserDB(db *bolt.DB) UserOpt {
	return func(u *User) {
		u.db = db
	}
}

func UserPassword(pw string) UserOpt {
	return func(u *User) {
		u.Password = pw
	}
}

func UserPermission(perm string) UserOpt {
	return func(u *User) {
		u.Permission = perm
	}
}

func UserTFA(tfa TFAer) UserOpt {
	return func(u *User) {
		u.tfa = tfa
	}
}

func NewUser(username string, opts ...UserOpt) *User {
	u := &User{Username: username}
	for _, o := range opts {
		o(u)
	}
	return u
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
			u.TFAData = []byte{}
			u.TFA = ""
			u.db = db
			users = append(users, u)
		}
		return nil
	})
	return users, err
}

func (u *User) SetDB(db *bolt.DB) {
	u.db = db
}

func (u *User) SetTFA(tfa TFAer) {
	u.tfa = tfa
}

func (u *User) Fetch() error {
	return u.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		v := b.Get([]byte(u.Username))
		if len(v) == 0 {
			return NotFound
		}
		return json.Unmarshal(v, u)
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

func (u *User) UpdateTFA() ([]byte, error) {
	if u.tfa == nil {
		return nil, errors.New("can't save user, no tfa")
	}

	var qr []byte
	var err error
	u.TFAData, qr, err = u.tfa.Get(u.Username)
	return qr, err
}

func (u *User) saveTFAData(savedUser *User) ([]byte, error) {
	//new user
	if savedUser == nil && u.TFAData == nil {
		if u.tfa == nil {
			return nil, errors.New("can't save user, no tfa")
		}
		var err error
		var qr []byte
		u.TFAData, qr, err = u.tfa.Get(u.Username)
		if err != nil {
			return nil, err
		}
		return qr, nil
	}

	//existing user
	if savedUser != nil && u.TFAData == nil {
		u.TFAData = savedUser.TFAData
	}
	return nil, nil
}

func (u *User) savePassword(savedUser *User) error {

	//new user
	if savedUser == nil && len(u.Password) == 0 {
		return errors.New("password is too short")
	}

	// existing user with no password update
	if savedUser != nil && len(u.Password) == 0 {
		u.HashedPassword = savedUser.HashedPassword
		return nil
	}
	u.hashPassword()
	return nil
}

func (u *User) Save() ([]byte, error) {
	savedUser := &User{Username: u.Username, db: u.db}
	if err := savedUser.Fetch(); err != nil && err != NotFound {
		return nil, err
	} else if err == NotFound {
		savedUser = nil
	}

	qr, err := u.saveTFAData(savedUser)
	if err != nil {
		return nil, err
	}

	if err := u.savePassword(savedUser); err != nil {
		return nil, err
	}

	return qr, u.db.Update(func(tx *bolt.Tx) error {
		d, _ := json.Marshal(u)
		b := tx.Bucket([]byte("users"))
		return b.Put([]byte(u.Username), d)
	})
}

func (u *User) Exists() bool {
	err := u.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		v := b.Get([]byte(u.Username))
		if len(v) == 0 {
			return NotFound
		}
		return json.Unmarshal(v, u)
	})
	if err == NotFound {
		return false
	}
	return err == nil
}

func (u *User) Delete() error {
	return u.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		return b.Delete([]byte(u.Username))
	})
}

func (u *User) CheckPassword() (bool, error) {
	pw := u.Password
	if len(u.HashedPassword) == 0 || len(u.TFAData) == 0 {
		if err := u.Fetch(); err != nil {
			LG.Println(err)
			return false, err
		}
	}

	if err := u.tfa.Check(u.TFAData, u.TFA); err != nil {
		LG.Println("tfa error", err)
		time.Sleep(500 * time.Millisecond)
		return false, nil
	}
	return bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(pw)) == nil, nil
}

func (u *User) hashPassword() {
	if len(u.Password) == 0 {
		return
	}
	u.HashedPassword, _ = bcrypt.GenerateFromPassword(
		[]byte(u.Password),
		bcrypt.DefaultCost,
	)
	u.Password = ""
}
