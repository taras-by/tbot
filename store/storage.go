package store

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

const (
	participantsBucketName = "participants"
	DBPath = "my.db"
)

type storage struct {
	db *bolt.DB
}

func NewStorage() (*storage, error) {

	bdb, err := bolt.Open(DBPath, 0600, nil)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to make boltdb for %s", DBPath)
	}

	bdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(participantsBucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &storage{
		db: bdb,
	}, nil
}

func (s *storage) Create(participant Participant) {
	s.save(participantsBucketName, participant.User.ID, participant)
}

func (s *storage) FindAll() (participants []Participant) {
	values := s.list(participantsBucketName)
	participants = []Participant{}
	for _, v := range values {
		participant := Participant{}
		_ = json.Unmarshal(v, &participant)
		participants = append(participants, participant)
	}
	return participants
}

func (s *storage) list(bucketName string) (values [][]byte) {
	values = [][]byte{}
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		b.ForEach(func(k, v []byte) error {
			values = append(values, v)
			return nil
		})
		return nil
	})
	return values
}

func (s *storage) save(bucketName string, key string, value interface{}) () {
	jsonData, _ := json.Marshal(value)
	s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		err := b.Put([]byte(key), []byte(jsonData))
		return err
	})
}

func (s *storage) Close() () {
	s.db.Close()
}
