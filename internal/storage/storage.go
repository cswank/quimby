package storage

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/asdine/storm"
	"github.com/cswank/quimby/internal/schema"
)

var db *storm.DB

func Get() *storm.DB {
	return db
}

func Close() {
	db.Close()
}

func init() {
	pth, err := dbPath()
	if err != nil {
		log.Fatalf("could not get database path: %s", err)
	}

	db, err = storm.Open(pth)
	if err != nil {
		log.Fatalf("could not open database: %s", err)
	}

	if err := db.Init(&schema.Gadget{}); err != nil {
		log.Fatalf("could not Init db: %v", err)
	}

	if err := db.Init(&schema.User{}); err != nil {
		log.Fatalf("could not Init db: %v", err)
	}
}

func dbPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(usr.HomeDir, ".quimby")

	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, 0777)
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(dir, "storm"), nil
}
