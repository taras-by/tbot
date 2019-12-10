package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/taras-by/tbot/store"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

const (
	telegramTokenFile = "telegram_token"
)

var (
	bot      *tgbotapi.BotAPI
	storage  *store.Storage
	commands = map[string]func(*tgbotapi.Message){
		"list":  list,
		"add":   add,
		"rm":    rm,
		"start": help,
		"help":  help,
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

		if update.Message.IsCommand() == false { // ignore any non-Command Updates
			continue
		}

		for command, funct := range commands {
			if update.Message.Command() == command {
				log.Printf("Command: \"%v\", arguments: \"%v\"", update.Message.Command(), update.Message.CommandArguments())
				funct(update.Message)
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
	participants := storage.FindByChatId(chatId)
	if len(participants) == 0 {
		return "No participants"
	}
	text = "List of participants:\n"
	for i, p := range participants {
		text = text + fmt.Sprintf(" *%v)* %v\n", i+1, p.User.Name)
	}
	return text
}

func rm(message *tgbotapi.Message) {
	chatId := message.Chat.ID
	participant, err := storage.Find(strconv.Itoa(message.From.ID), message.Chat.ID)
	if err != nil {
		sendMessageToChat(chatId, "You are not a participant yet")
		return
	}

	storage.Delete(participant)
	sendMessageToChat(message.Chat.ID, fmt.Sprintf("Removed *%s*", participant.User.Name))

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
