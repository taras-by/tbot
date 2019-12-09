package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/taras-by/tbot/store"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	telegramTokenFile = "telegram_token"
)

var (
	bot      *tgbotapi.BotAPI
	storage  *store.Storage
	commands = map[string]func(*tgbotapi.Message){
		"/list":  list,
		"/add":   add,
		"/+":     add,
		"/rm":    rm,
		"/-":     rm,
		"/start": help,
		"/help":  help,
	}
)

func server() (err error) {
	token, err := ioutil.ReadFile(telegramTokenFile)
	if err != nil {
		log.Panic(err.Error())
	}

	bot, err = tgbotapi.NewBotAPI(string(token))
	if err != nil {
		log.Panic(err.Error())
	}

	storage, err = store.NewStorage()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("Message: [%s] %s", update.Message.From.UserName, update.Message.Text)

		for command, funct := range commands {
			text := update.Message.Text
			if strings.HasPrefix(strings.ToLower(text), command) {
				funct(update.Message)
				log.Printf("Command: %v", command)
			}
		}

	}
	return nil
}

func list(message *tgbotapi.Message) {
	chatId := message.Chat.ID
	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func add(message *tgbotapi.Message) {
	chatId := message.Chat.ID

	storage.Create(
		store.Participant{
			User: store.User{
				ID:   strconv.Itoa(message.From.ID),
				Name: message.From.UserName,
			},
			Time:   time.Now(),
			ChatId: chatId,
		},
	)

	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func participantsText(chatId int64) (text string) {
	text = "List of participants:\n"
	for i, p := range storage.FindByChatId(chatId) {
		text = text + fmt.Sprintf(" *%v)* %v\n", i+1, p.User.Name)
	}
	return text
}

func rm(message *tgbotapi.Message) {
	sendMessageToChat(message.Chat.ID, "Removed ...")
	chatId := message.Chat.ID
	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func help(message *tgbotapi.Message) {
	sendMessageToChat(message.Chat.ID, "Help: ...")
}

func sendMessageToChat(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "markdown"
	_, _ = bot.Send(msg)
}
