package quimby

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/securecookie"
)

var (
	DB          *bolt.DB
	privKey     *rsa.PrivateKey
	PubKey      *rsa.PublicKey
	pubKeyPath  = os.Getenv("QUIMBY_JWT_PUB")
	privKeyPath = os.Getenv("QUIMBY_JWT_PRIV")

	hashKey  = []byte(os.Getenv("QUIMBY_HASH_KEY"))
	blockKey = []byte(os.Getenv("QUIMBY_BLOCK_KEY"))
	SC       = securecookie.New(hashKey, blockKey)
)

func GenerateCookie(username string) *http.Cookie {
	value := map[string]string{
		"user": username,
	}

	encoded, _ := SC.Encode("quimby", value)
	return &http.Cookie{
		Name:     "quimby",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
	}
}

func GetUserFromCookie(r *http.Request) (*User, error) {
	user := &User{
		DB: DB,
	}
	cookie, err := r.Cookie("quimby")
	if err != nil {
		return nil, err
	}
	var m map[string]string
	err = SC.Decode("quimby", cookie.Value, &m)
	if err != nil {
		return nil, err
	}
	if m["user"] == "" {
		return nil, errors.New("no way, eh")
	}
	user.Username = m["user"]
	err = user.Fetch()
	user.HashedPassword = []byte{}
	return user, err
}

func GenerateToken(username string, d time.Duration) (string, error) {
	if PubKey == nil || privKey == nil {
		PubKey = getPublicKey()
		privKey = getPrivateKey()
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodRS512,
		jwt.MapClaims{
			"exp": time.Now().Add(d).Unix(),
			"iat": time.Now().Unix(),
			"sub": username,
		},
	)
	tokenString, err := token.SignedString(privKey)
	if err != nil {
		panic(err)
	}
	return tokenString, nil
}

func GetUserFromToken(r *http.Request) (*User, error) {
	if PubKey == nil || privKey == nil {
		PubKey = getPublicKey()
		privKey = getPrivateKey()
	}

	user := &User{}
	token, err := jwt.Parse(r.Header.Get("Authorization"), func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return PubKey, nil
	})
	if err != nil {
		return nil, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user.Username = claims["sub"].(string)
	} else {
		return nil, errors.New("invalid token")
	}

	return user, nil
}

func getPrivateKey() *rsa.PrivateKey {
	privateKeyFile, err := os.Open(privKeyPath)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := privateKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))
	privateKeyFile.Close()
	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	if err != nil {
		panic(err)
	}
	return privateKeyImported
}

func getPublicKey() *rsa.PublicKey {
	publicKeyFile, err := os.Open(pubKeyPath)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := publicKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	publicKeyFile.Close()
	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)
	if err != nil {
		panic(err)
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)
	if !ok {
		panic(err)
	}

	return rsaPub
}
