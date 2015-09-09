package models

import "github.com/boltdb/bolt"

func GetDB(pth string) (*bolt.DB, error) {
	db, err := bolt.Open(pth, 0600, nil)
	if err != nil {
		return nil, err
	}

	return db, db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("gadgets"))
		return err
	})
}
