package store

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type storage struct {
	db *bolt.DB
}

func NewStorage() (*storage, error) {

	fileName := "my.db"
	bdb, err := bolt.Open(fileName, 0600, nil)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to make boltdb for %s", fileName)
	}

	bdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &storage{
		db: bdb,
	}, nil
}

func (s *storage) List() () {
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("%s: %s\n", k, v)
			return nil
		})
		return nil
	})
}

func (s *storage) Insert(key string, value string) () {
	s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		err := b.Put([]byte(key), []byte(value))
		return err
	})
}

func (s *storage) Close() () {
	s.db.Close()
}
