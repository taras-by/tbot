package main

import (
	"github.com/bxcodec/faker"
	"github.com/satori/go.uuid"
	"github.com/taras-by/tbot/store"
	"log"
)

func main() {
	uid := uuid.Must(uuid.NewV4()).String()
	text := faker.Sentence()

	storage, err := store.NewStorage()
	if err != nil {
		log.Fatal(err)
	}

	storage.Insert(uid, text)
	storage.List()
	defer storage.Close()
}
