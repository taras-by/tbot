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
		"list":  list,
		"add":   add,
		"rm":    rm,
		"reset": reset,
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
	var participant store.Participant

	if message.CommandArguments() == "" {
		participant = storage.Create(
			store.Participant{
				User: store.User{
					Id:   strconv.Itoa(message.From.ID),
					Name: message.From.UserName,
					Type: store.UserTelegram,
				},
				Time:   time.Now(),
				ChatId: chatId,
			},
		)
	} else {
		participant = storage.Create(
			store.Participant{
				User: store.User{
					Name: store.Escape(message.CommandArguments()),
					Type: store.UserGuest,
				},
				Time:   time.Now(),
				ChatId: chatId,
			},
		)
	}
	sendMessageToChat(chatId, fmt.Sprintf("Added *%s*", participant.Link()))
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
		text = text + fmt.Sprintf(" *%v)* %v\n", i+1, p.Name())
	}
	return text
}

func rm(message *tgbotapi.Message) {
	var participant store.Participant
	var err error
	chatId := message.Chat.ID
	args := strings.TrimSpace(message.CommandArguments())

	if args == "" {
		participant, err = storage.Find(strconv.Itoa(message.From.ID), chatId)
		if err != nil {
			sendMessageToChat(chatId, "You are not a participant yet")
			return
		}
	} else {
		participant, err = storage.FindByName(args, chatId)
		if err != nil {
			sendMessageToChat(chatId, err.Error())
			return
		}
	}

	storage.Delete(participant)
	sendMessageToChat(chatId, fmt.Sprintf("Removed *%s*", participant.Link()))

	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func reset(message *tgbotapi.Message) {
	chatId := message.Chat.ID
	err := storage.DeleteAll(chatId)
	if err != nil {
		sendMessageToChat(chatId, err.Error())
		return
	}

	sendMessageToChat(chatId, "All participants was deleted")
}

func help(message *tgbotapi.Message) {
	sendMessageToChat(message.Chat.ID, "Help: ...")
}

func sendMessageToChat(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "markdown"
	_, _ = bot.Send(msg)
}
