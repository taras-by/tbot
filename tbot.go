package main

import (
	"fmt"
	"github.com/bxcodec/faker"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/satori/go.uuid"
	"github.com/taras-by/tbot/store"
	"io/ioutil"
	"log"
	"time"
)

const (
	telegramTokenFile = "telegram_token"
)

func main() {

	storage, err := store.NewStorage()
	if err != nil {
		log.Fatal(err)
	}

	token, err := ioutil.ReadFile(telegramTokenFile)
	if err != nil {
		log.Panic(err.Error())
	}

	bot, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		log.Panic(err.Error())
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

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
