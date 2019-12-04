package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func main() {
	insert("key1", "value 1")
	view()
}
func view() {

	db, err := bolt.Open("my.db", 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("%s: %s\n", k, v)
			return nil
		})
		return nil
	})

	defer db.Close()
}

func insert(key string, value string) {
	db, err := bolt.Open("my.db", 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		err := b.Put([]byte(key), []byte(value))
		return err
	})

	defer db.Close()
}

func create() {
	db, err := bolt.Open("my.db", 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	defer db.Close()
}
