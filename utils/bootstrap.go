package utils

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
)

var (
	template = `export QUIMBY_DB="{{.DBPath}}"
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

func Bootstrap() {
	var b bootstrap
	db := getDB(&b)

	saveQuimbyUser(db)
	AddUser(db)

}

func saveQuimbyUser(db *bolt.DB) {

}

func getDB(b *bootstrap) *bolt.DB {
	fmt.Printf("data directory path: ")
	fmt.Scanf("%s", &b.Home)

	b.DBPath = filepath.Join(b.Home, "quimby.db")
	db, err := models.GetDB(b.DBPath)
	if err != nil {
		log.Fatalf("could not open db at %s - %v", b.DBPath, err)
	}
	return db
}
