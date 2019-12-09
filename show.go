package main

import (
	"fmt"
	"github.com/taras-by/tbot/store"
	"log"
)

func show() (err error) {

	storage, err := store.NewStorage()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	for _, p := range storage.FindAll() {
		fmt.Printf("Participant: %v\n", p)
	}

	return nil
}
