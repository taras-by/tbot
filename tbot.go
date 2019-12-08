package main

import (
	"fmt"
	"github.com/bxcodec/faker"
	"github.com/satori/go.uuid"
	"github.com/taras-by/tbot/store"
	"log"
	"time"
)

func main() {

	storage, err := store.NewStorage()
	if err != nil {
		log.Fatal(err)
	}

	storage.Create(store.Participant{
		User: store.User{
			ID:   uuid.Must(uuid.NewV4()).String(),
			Name: faker.Name(),
		},
		Time: time.Now(),
	})

	for _, p := range storage.FindAll() {
		fmt.Printf("Participant: %v\n", p)
	}

	defer storage.Close()
}
