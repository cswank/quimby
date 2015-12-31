package utils

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	env         = `export QUIMBY_DB="{{.DBPath}}"
export QUIMBY_INTERFACE="0.0.0.0"
export QUIMBY_PORT=443
export QUIMBY_INTERNAL_PORT=8989
export QUIMBY_HOST="http://{{.IP}}"
export QUIMBY_BLOCK_KEY="{{.Block}}"
export QUIMBY_HASH_KEY= "{{.Hash}}"
export QUIMBY_JWT_PRIV="{{.JWTPriv}}"
export QUIMBY_JWT_PUB="{{.JWTPub}}"
export QUIMBY_TLS_KEY="{{.TLSKey}}"
export QUIMBY_TLS_CERT="{{.TLSCert}}"
export QUIMBY_USER="quimby"
`
)

type bootstrap struct {
	Home    string
	DBPath  string
	IP      string
	Block   string
	Hash    string
	JWTPriv string
	JWTPub  string
	TLSKey  string
	TLSCert string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Bootstrap() {
	var b bootstrap
	db := getDB(&b)
	saveQuimbyUser(db)
	AddUser(db)
	addCert(&b, db)
	addIP(&b)
	writeEnv(&b)
}

func addIP(b *bootstrap) {
	fmt.Printf("local ip address of server: ")
	fmt.Scanf("%s", &b.IP)
}

func addCert(b *bootstrap, db *bolt.DB) {
	pth := filepath.Join(b.Home, "certs", "jwt")
	b.TLSKey = filepath.Join(b.Home, "certs", "tls", "key.pem")
	b.TLSCert = filepath.Join(b.Home, "certs", "tls", "cert.pem")

	if err := os.MkdirAll(pth, 0777); err != nil {
		log.Fatal(err)
	}
	var dom string
	fmt.Printf("public web domain (like example.com): ")
	fmt.Scanf("%s", &dom)
	GenerateCert(dom, pth)
	writeEnv(b)
}

func writeEnv(b *bootstrap) {
	b.Block = randString(16)
	b.Hash = randString(16)
	b.JWTPriv = filepath.Join(b.Home, "certs", "jwt", "private_key.pem")
	b.JWTPub = filepath.Join(b.Home, "certs", "jwt", "public_key.pem")
	p := filepath.Join(b.Home, "quimby.env")
	t := template.Must(template.New("env").Parse(env))
	f, err := os.Create(p)
	if err != nil {
		log.Fatal(err)
	}
	if err := t.Execute(f, b); err != nil {
		if err != nil {
			log.Fatal(err)
		}
	}
	f.Close()
}

func saveQuimbyUser(db *bolt.DB) {
	u := models.User{
		DB:         db,
		Username:   "quimby",
		Password:   randString(16),
		Permission: "write",
	}
	if err := u.Save(); err != nil {
		log.Fatal(err)
	}
}

func getDB(b *bootstrap) *bolt.DB {
	fmt.Printf("data directory path: ")
	fmt.Scanf("%s", &b.Home)

	if err := os.MkdirAll(b.Home, 0777); err != nil {
		log.Fatal(err)
	}

	b.DBPath = filepath.Join(b.Home, "quimby.db")
	db, err := models.GetDB(b.DBPath)
	if err != nil {
		log.Fatalf("could not open db at %s - %v", b.DBPath, err)
	}
	return db
}

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
