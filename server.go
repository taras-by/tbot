package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/taras-by/tbot/store"
	"log"
	"regexp"
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
		"ping":  ping,
		"start": help,
		"help":  help,
	}
	linkArgsChecker    = regexp.MustCompile(`^@(\S+)$`)
	integerArgsChecker = regexp.MustCompile(`^\d+$`)
)

func server() (err error) {

	bot, err = tgbotapi.NewBotAPI(Opts.TelegramToken)
	if err != nil {
		log.Printf("Telegram connection Error. Token: %s", Opts.TelegramToken)
		log.Panic(err.Error())
	}

	storage, err = store.NewStorage(Opts.StorePath)
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
	args := strings.TrimSpace(message.CommandArguments())

	if args == "" {
		creationTime := time.Now()
		existingParticipant, err := storage.FindByLink("@"+message.From.UserName, chatId)
		if err == nil {
			if existingParticipant.IsUnresolved() == true {
				creationTime = existingParticipant.Time
				storage.Delete(existingParticipant)
				//sendMessageToChat(chatId, "Update unresolved")
			} else {
				sendMessageToChat(chatId, "You are already a participant")
				return
			}
		}
		participant = storage.Create(
			store.Participant{
				User: store.User{
					Id:        strconv.Itoa(message.From.ID),
					UserName:  message.From.UserName,
					FirstName: message.From.FirstName,
					LastName:  message.From.LastName,
					Type:      store.UserTelegram,
				},
				Time:   creationTime,
				ChatId: chatId,
			},
		)
	} else if integerArgsChecker.MatchString(args) {
		sendMessageToChat(chatId, "Fail. UserName as an number")
		return
	} else if match := linkArgsChecker.FindStringSubmatch(args); len(match) == 2 {
		userName := match[1]
		existingParticipant, err := storage.FindByLink("@"+userName, chatId)
		if err == nil && existingParticipant.Id() != "" {
			sendMessageToChat(chatId, "User is already in the list of participants")
			return
		}
		participant = storage.Create(
			store.Participant{
				User: store.User{
					UserName: userName,
					Type:     store.UserUnresolved,
				},
				Time:   time.Now(),
				ChatId: chatId,
			},
		)
	} else {
		participant = storage.Create(
			store.Participant{
				User: store.User{
					UserName: args,
					Type:     store.UserGuest,
				},
				Time:   time.Now(),
				ChatId: chatId,
			},
		)
	}
	sendMessageToChat(chatId, fmt.Sprintf("Added %s", store.Escape(participant.Link())))
	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func rm(message *tgbotapi.Message) {
	var participant store.Participant
	var err error
	chatId := message.Chat.ID
	args := strings.TrimSpace(message.CommandArguments())

	if args == "" {
		participant, err = storage.Find(strconv.Itoa(message.From.ID), chatId)
		if err != nil {
			participant, err = storage.FindByLink("@"+message.From.UserName, chatId)
			if err != nil {
				sendMessageToChat(chatId, "You are not a participant yet")
				return
			}
		}
	} else if linkArgsChecker.MatchString(args) {
		linkString := string(linkArgsChecker.Find([]byte(args)))
		participant, err = storage.FindByLink(linkString, chatId)
		if err != nil {
			sendMessageToChat(chatId, store.Escape(err.Error()))
			return
		}
	} else if integerArgsChecker.MatchString(args) {
		numberString := string(integerArgsChecker.Find([]byte(args)))
		number, err := strconv.Atoi(numberString)
		if err != nil {
			sendMessageToChat(chatId, store.Escape(err.Error()))
			return
		}
		participant, err = storage.FindByNumber(number, chatId)
		if err != nil {
			sendMessageToChat(chatId, store.Escape(err.Error()))
			return
		}
	} else {
		participant, err = storage.FindByName(args, chatId)
		if err != nil {
			sendMessageToChat(chatId, store.Escape(err.Error()))
			return
		}
	}

	storage.Delete(participant)
	sendMessageToChat(chatId, fmt.Sprintf("Removed %s", store.Escape(participant.Link())))

	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func reset(message *tgbotapi.Message) {
	chatId := message.Chat.ID
	err := storage.DeleteAll(chatId)
	if err != nil {
		sendMessageToChat(chatId, store.Escape(err.Error()))
		return
	}

	sendMessageToChat(chatId, "All participants was deleted")
}

func participantsText(chatId int64) (text string) {
	participants := storage.FindByChatId(chatId)
	if len(participants) == 0 {
		return "No participants"
	}
	text = "List of participants:\n"
	for i, p := range participants {
		text = text + fmt.Sprintf(" *%v)* %v\n", i+1, store.Escape(p.Name()))
	}
	return text
}

func help(message *tgbotapi.Message) {
	text := "*Help:*\n" +
		"/list - participants list\n" +
		"/add - add yourself or someone\n" +
		"/rm - remove yourself or someone\n" +
		"/reset - remove all\n" +
		//"/ping - turn to non-participants\n" +
		"/help - help\n" +
		"\n" +
		"*Examples:*\n" +
		"``` /add @smith\n" +
		" /add My brother John\n" +
		" /rm @smith\n" +
		" /rm My brother John\n" +
		" /rm 3\n" +
		"```\n" +
		"The last example is the removal of the third participant\n"
	sendMessageToChat(message.Chat.ID, text)
}

func ping(message *tgbotapi.Message) {
	sendMessageToChat(message.Chat.ID, "Turn to non-participants... *Not implemented*.\nWelcome to https://github.com/taras-by/tbot")
}

func sendMessageToChat(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "markdown"
	_, err := bot.Send(msg)
	if err != nil {
		log.Print(err)
		log.Print(text)
	}
}
