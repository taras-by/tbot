package store

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"log"
)

const (
	participantsBucketName = "participants"
	DBPath                 = "my.db"
)

type Storage struct {
	db *bolt.DB
}

func NewStorage() (*Storage, error) {

	bdb, err := bolt.Open(DBPath, 0600, nil)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to make boltdb for %s", DBPath)
	}
	log.Printf("Storage opened")

	bdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(participantsBucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &Storage{
		db: bdb,
	}, nil
}

func (s *Storage) Create(participant Participant) {
	chatParticipantsBucketName := chatParticipantsBucketName(participant.ChatId)
	s.save(chatParticipantsBucketName, participant.User.ID, participant)
}

func (s *Storage) FindAll() (participants []Participant) {
	values := s.list(participantsBucketName)
	participants = []Participant{}
	for _, v := range values {
		participant := Participant{}
		_ = json.Unmarshal(v, &participant)
		participants = append(participants, participant)
	}
	return participants
}

func (s *Storage) FindByChatId(chatId int64) (participants []Participant) {
	chatParticipantsBucketName := chatParticipantsBucketName(chatId)
	values := s.list(chatParticipantsBucketName)
	participants = []Participant{}
	for _, v := range values {
		participant := Participant{}
		_ = json.Unmarshal(v, &participant)
		participants = append(participants, participant)
	}
	return participants
}

func chatParticipantsBucketName(chatId int64) string {
	return string(chatId) + participantsBucketName
}

func (s *Storage) list(bucketName string) (values [][]byte) {
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

func (s *Storage) save(bucketName string, key string, value interface{}) () {

	s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	jsonData, _ := json.Marshal(value)
	s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		err := b.Put([]byte(key), []byte(jsonData))
		return err
	})
}

func (s *Storage) Close() () {
	_=s.db.Close()
	log.Printf("Storage closed")
}
