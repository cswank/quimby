package controllers

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

	"github.com/cswank/quimby/models"
	jwt "github.com/dgrijalva/jwt-go"
)

var (
	privKey     *rsa.PrivateKey
	PubKey      *rsa.PublicKey
	pubKeyPath  = os.Getenv("QUIMBY_PUB_KEY")
	privKeyPath = os.Getenv("QUIMBY_PRIV_KEY")
)

const (
	tokenDuration = 72
	expireOffset  = 3600
)

func init() {
	PubKey = getPublicKey()
	privKey = getPrivateKey()
}

func getUserFromToken(r *http.Request) (*models.User, error) {
	user := &models.User{
		DB: DB,
	}

	token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			return PubKey, nil
		}
	})

	if err != nil || !token.Valid {
		return nil, errors.New("no way, eh")
	}

	user.Username = token.Claims["sub"].(string)
	err = user.Fetch()
	user.HashedPassword = []byte{}
	return user, err
}

func generateToken(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodRS512)
	token.Claims["exp"] = time.Now().Add(time.Duration(24 * time.Hour)).Unix()
	token.Claims["iat"] = time.Now().Unix()
	token.Claims["sub"] = user.Username
	tokenString, err := token.SignedString(privKey)
	if err != nil {
		panic(err)
	}
	return tokenString, nil
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
