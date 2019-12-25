package store

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"log"
	"sort"
	"strconv"
)

const (
	chatsBucketName = "chats"
)

type Storage struct {
	db *bolt.DB
}

func NewStorage(storePath string) (*Storage, error) {

	bdb, err := bolt.Open(storePath, 0600, nil)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to make boltdb for %s", storePath)
	}
	log.Printf("Storage opened in %s", storePath)

	bdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(chatsBucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &Storage{
		db: bdb,
	}, nil
}

func (s *Storage) Close() () {
	_ = s.db.Close()
	log.Print("Storage closed")
}

func (s *Storage) Create(participant Participant) Participant {
	_ = s.db.Update(func(tx *bolt.Tx) (err error) {
		var chatBkt *bolt.Bucket

		if chatBkt, err = s.makeChatBucket(tx, participant.ChatId); err != nil {
			return err
		}

		if err = s.save(chatBkt, participant.Id(), participant); err != nil {
			return errors.Wrapf(err, "failed to put key %s to bucket %s", participant.Id(), participant)
		}

		return nil
	})

	return participant
}

func (s *Storage) Delete(participant Participant) {
	_ = s.db.Update(func(tx *bolt.Tx) (err error) {
		var chatBkt *bolt.Bucket

		if chatBkt, err = s.makeChatBucket(tx, participant.ChatId); err != nil {
			return err
		}

		if err = chatBkt.Delete([]byte(participant.Id())); err != nil {
			return errors.Wrapf(err, "failed to delete key %s from chat bucket %v", participant.Id(), participant.ChatId)
		}

		return nil
	})
}

func (s *Storage) DeleteAll(chatId int64) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		chatsBkt := tx.Bucket([]byte(chatsBucketName))
		chatBucketName := strconv.FormatInt(chatId, 10)
		if e := chatsBkt.DeleteBucket([]byte(chatBucketName)); e != nil {
			return errors.Wrapf(e, "Failed to delete participants")
		}
		return nil
	})
	return errors.Wrapf(err, "Failed to delete participants")
}

func (s *Storage) Find(key string, chatId int64) (participant Participant, err error) {
	err = s.db.View(func(tx *bolt.Tx) (err error) {
		var chatBkt *bolt.Bucket

		if chatBkt, err = s.getChatBucket(tx, chatId); err != nil {
			return err
		}

		value := chatBkt.Get([]byte(key))
		if value == nil {
			return errors.Errorf("no value for %s", key)
		}

		if err := json.Unmarshal(value, &participant); err != nil {
			return errors.Wrap(err, "failed to unmarshal")
		}

		return nil
	})
	return participant, err
}

func (s *Storage) FindByNumber(number int, chatId int64) (participant Participant, err error) {
	participants := s.FindByChatId(chatId)
	for i, p := range participants {
		if number == i+1 {
			return p, nil
		}
	}
	return participant, errors.Errorf("Participant with number %d not found", number)
}

func (s *Storage) FindByName(name string, chatId int64) (participant Participant, err error) {
	participants := s.FindByChatId(chatId)
	for _, p := range participants {
		if p.Name() == name {
			return p, nil
		}
	}
	return participant, errors.Errorf("Participant with name \"%s\" not found", name)
}

func (s *Storage) FindByLink(name string, chatId int64) (participant Participant, err error) {
	participants := s.FindByChatId(chatId)
	for _, p := range participants {
		if p.Link() == name {
			return p, nil
		}
	}
	return participant, errors.Errorf("Participant with link %s not found", name)
}

func (s *Storage) FindAll() (participants []Participant) {
	values := s.list(chatsBucketName)
	participants = []Participant{}
	for _, v := range values {
		participant := Participant{}
		e := json.Unmarshal(v, &participant)
		if e != nil {
			log.Printf("Error")
		}
		participants = append(participants, participant)
	}
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Time.After(participants[j].Time)
	})
	return participants
}

func (s *Storage) FindByChatId(chatId int64) (participants []Participant) {

	_ = s.db.View(func(tx *bolt.Tx) error {

		bucket, e := s.getChatBucket(tx, chatId)
		if e != nil {
			return e
		}
		bucket.ForEach(func(k, v []byte) error {
			participant := Participant{}
			if e = json.Unmarshal(v, &participant); e != nil {
				return errors.Wrap(e, "failed to unmarshal")
			}
			participants = append(participants, participant)
			return nil
		})
		return nil
	})
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Time.Before(participants[j].Time)
	})
	return participants
}

func (s *Storage) CountByChatId(chatId int64) int {
	participants := s.FindByChatId(chatId)
	return len(participants)
}

func (s *Storage) list(bucketName string) (values [][]byte) {
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		b.ForEach(func(k, v []byte) error {
			b.Bucket(k).ForEach(func(k, v []byte) error {
				values = append(values, v)
				return nil
			})
			return nil
		})
		return nil
	})
	return values
}

func (s *Storage) save(bkt *bolt.Bucket, key string, value interface{}) (err error) {
	jsonData, _ := json.Marshal(value)
	err = bkt.Put([]byte(key), []byte(jsonData))
	if err != nil {
		return errors.Wrapf(err, "Error")
	}
	return nil
}

func (s *Storage) getChatBucket(tx *bolt.Tx, chatId int64) (*bolt.Bucket, error) {
	chatBkt := tx.Bucket([]byte(chatsBucketName))
	if chatBkt == nil {
		return nil, errors.Errorf("no bucket %s", chatsBucketName)
	}
	res := chatBkt.Bucket([]byte(strconv.FormatInt(chatId, 10)))
	if res == nil {
		return nil, errors.Errorf("no bucket %s in store", chatId)
	}
	return res, nil
}

func (s *Storage) makeChatBucket(tx *bolt.Tx, chatId int64) (*bolt.Bucket, error) {
	chatsBkt := tx.Bucket([]byte(chatsBucketName))
	if chatsBkt == nil {
		return nil, errors.Errorf("no bucket %s", chatsBucketName)
	}
	chatBucketName := strconv.FormatInt(chatId, 10)
	res, err := chatsBkt.CreateBucketIfNotExists([]byte(chatBucketName))
	if err != nil {
		return nil, errors.Wrapf(err, "no bucket %s in store", chatId)
	}
	return res, nil
}
