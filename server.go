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
	maxLengthStringArgument = 50
)

var (
	bot     *tgbotapi.BotAPI
	storage *store.Storage
	routes  = []route{
		{`add`, ``, addMe},
		{`add`, `^@(\S+)$`, addByLink},
		{`add`, `^\d+$`, addByNumber},
		{`add`, `^[\w\s\.,]+$`, addByName},
		{`rm`, ``, removeMe},
		{`rm`, `^@(\S+)$`, removeByLink},
		{`rm`, `^\d+$`, removeByNumber},
		{`rm`, `^[\w\s\.,]+$`, removeByName},
		{`list`, ``, list},
		{`ping`, ``, ping},
		{`reset`, ``, reset},
		{`start`, ``, help},
		{`help`, ``, help},
	}
)

type route struct {
	BotCommand    string
	ArgExpression string
	Command       func(int64, string, *regexp.Regexp, *tgbotapi.Message)
}

func server() (err error) {

	printVersion()

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

		args := strings.TrimSpace(update.Message.CommandArguments())
		cmd := update.Message.Command()
		chatId := update.Message.Chat.ID

		if len([]rune(args)) > maxLengthStringArgument {
			sendMessageToChat(chatId, store.Escape("Parameter too long"))
			continue
		}

		commandIsOk := false
		for _, route := range routes {

			checker := regexp.MustCompile(route.ArgExpression)
			argsIsMatched := false

			if route.ArgExpression != "" && checker.MatchString(args) {
				argsIsMatched = true
			}

			if route.ArgExpression == "" && args == "" {
				argsIsMatched = true
			}

			if route.BotCommand == cmd && argsIsMatched {
				commandIsOk = true
				log.Printf("Command: \"%v\", arguments: \"%v\"", cmd, args)
				route.Command(chatId, args, checker, update.Message)
				break
			}
		}
		if commandIsOk == false {
			sendMessageToChat(chatId, store.Escape("Wrong command"))
		}
	}
	return nil
}

func list(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {
	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func addMe(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {

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

	participant := storage.Create(
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

	sendMessageToChat(chatId, fmt.Sprintf("Added %s", store.Escape(participant.Link())))
	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func addByLink(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {

	match := checker.FindStringSubmatch(args)
	if len(match) != 2 {
		sendMessageToChat(chatId, "Error")
		return
	}

	userName := match[1]
	existingParticipant, err := storage.FindByLink("@"+userName, chatId)
	if err == nil && existingParticipant.Id() != "" {
		sendMessageToChat(chatId, "User is already in the list of participants")
		return
	}

	participant := storage.Create(
		store.Participant{
			User: store.User{
				UserName: userName,
				Type:     store.UserUnresolved,
			},
			Time:   time.Now(),
			ChatId: chatId,
		},
	)

	sendMessageToChat(chatId, fmt.Sprintf("Added %s", store.Escape(participant.Link())))
	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func addByName(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {

	participant := storage.Create(
		store.Participant{
			User: store.User{
				UserName: args,
				Type:     store.UserGuest,
			},
			Time:   time.Now(),
			ChatId: chatId,
		},
	)

	sendMessageToChat(chatId, fmt.Sprintf("Added %s", store.Escape(participant.Link())))
	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func addByNumber(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {
	sendMessageToChat(chatId, "Fail. UserName as an number")
}

func removeMe(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {

	var participant store.Participant
	var err error

	participant, err = storage.Find(strconv.Itoa(message.From.ID), chatId)
	if err != nil {
		participant, err = storage.FindByLink("@"+message.From.UserName, chatId)
		if err != nil {
			sendMessageToChat(chatId, "You are not a participant yet")
			return
		}
	}

	storage.Delete(participant)
	sendMessageToChat(chatId, fmt.Sprintf("Removed %s", store.Escape(participant.Link())))

	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func removeByNumber(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {

	var participant store.Participant
	var err error

	numberString := string(checker.Find([]byte(args)))
	number, err := strconv.Atoi(numberString)
	if err != nil {
		sendMessageToChat(chatId, "Wrong parameter")
		return
	}

	participant, err = storage.FindByNumber(number, chatId)
	if err != nil {
		sendMessageToChat(chatId, store.Escape(err.Error()))
		return
	}

	storage.Delete(participant)
	sendMessageToChat(chatId, fmt.Sprintf("Removed %s", store.Escape(participant.Link())))

	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func removeByLink(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {

	var participant store.Participant
	var err error

	linkString := string(checker.Find([]byte(args)))
	participant, err = storage.FindByLink(linkString, chatId)
	if err != nil {
		sendMessageToChat(chatId, store.Escape(err.Error()))
		return
	}

	storage.Delete(participant)
	sendMessageToChat(chatId, fmt.Sprintf("Removed %s", store.Escape(participant.Link())))

	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func removeByName(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {

	var participant store.Participant
	var err error

	participant, err = storage.FindByName(args, chatId)
	if err != nil {
		sendMessageToChat(chatId, store.Escape(err.Error()))
		return
	}

	storage.Delete(participant)
	sendMessageToChat(chatId, fmt.Sprintf("Removed %s", store.Escape(participant.Link())))

	text := participantsText(chatId)
	sendMessageToChat(chatId, text)
}

func reset(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {
	err := storage.DeleteAll(chatId)
	if err != nil {
		sendMessageToChat(chatId, store.Escape(err.Error()))
		return
	}

	sendMessageToChat(chatId, "All participants was deleted")
}

func help(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {
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
		"The last example is the removal of the third participant\n\n" +
		"_Version: " + Version + "_"
	sendMessageToChat(chatId, text)
}

func ping(chatId int64, args string, checker *regexp.Regexp, message *tgbotapi.Message) {
	sendMessageToChat(chatId, "Turn to non-participants... *Not implemented*.\nWelcome to https://github.com/taras-by/tbot")
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

func sendMessageToChat(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "markdown"
	_, err := bot.Send(msg)
	if err != nil {
		log.Print(err)
		log.Print(text)
	}
}
