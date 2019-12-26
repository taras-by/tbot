package telegram

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
	maxParticipants         = 100
)

type MessageHandler struct {
	Bot     *tgbotapi.BotAPI
	Storage *store.Storage
	routes  []route
	Version string
}

type route struct {
	botCommand    string
	argExpression string
	command       func(c conversation)
}

type conversation struct {
	chatId  int64
	args    string
	checker *regexp.Regexp
	message *tgbotapi.Message
}

func (h *MessageHandler) handle(message *tgbotapi.Message) {
	if message == nil { // ignore any non-Message Updates
		return
	}

	log.Printf("Message: [%s] %s", message.From.UserName, message.Text)

	if message.IsCommand() == false { // ignore any non-command Updates
		return
	}

	args := strings.TrimSpace(message.CommandArguments())
	cmd := message.Command()
	chatId := message.Chat.ID

	if len([]rune(args)) > maxLengthStringArgument {
		h.sendMessageToChat(chatId, store.Escape("Parameter too long"))
		return
	}

	commandIsOk := false
	for _, route := range h.routes {

		checker := regexp.MustCompile(route.argExpression)
		argsIsMatched := false

		if route.argExpression != "" && checker.MatchString(args) {
			argsIsMatched = true
		}

		if route.argExpression == "" && args == "" {
			argsIsMatched = true
		}

		if route.botCommand == cmd && argsIsMatched {
			commandIsOk = true
			log.Printf("command: \"%v\", arguments: \"%v\"", cmd, args)

			c := conversation{
				chatId:  chatId,
				args:    args,
				checker: checker,
				message: message,
			}

			route.command(c)
			break
		}
	}
	if commandIsOk == false {
		h.sendMessageToChat(chatId, store.Escape("Wrong command"))
		return
	}
}

func (h *MessageHandler) list(c conversation) {
	text := h.participantsText(c.chatId)
	h.sendMessageToChat(c.chatId, text)
}

func (h *MessageHandler) addMe(c conversation) {

	creationTime := time.Now()
	existingParticipant, err := h.Storage.FindByLink("@"+c.message.From.UserName, c.chatId)
	if err == nil {
		if existingParticipant.IsUnresolved() == true {
			creationTime = existingParticipant.Time
			h.Storage.Delete(existingParticipant)
			//sendMessageToChat(chatId, "Update unresolved")
		} else {
			h.sendMessageToChat(c.chatId, "You are already a participant")
			return
		}
	}

	if h.Storage.CountByChatId(c.chatId) >= maxParticipants {
		h.sendMessageToChat(c.chatId, fmt.Sprintf("Maximum chat participants: *%v*", maxParticipants))
		return
	}

	participant := h.Storage.Create(
		store.Participant{
			User: store.User{
				Id:        strconv.Itoa(c.message.From.ID),
				UserName:  c.message.From.UserName,
				FirstName: c.message.From.FirstName,
				LastName:  c.message.From.LastName,
				Type:      store.UserTelegram,
			},
			Time:   creationTime,
			ChatId: c.chatId,
		},
	)

	text := fmt.Sprintf("*Added* %s", store.Escape(participant.Link())) + "\n" +
		h.participantsText(c.chatId)
	h.sendMessageToChat(c.chatId, text)
}

func (h *MessageHandler) addByLink(c conversation) {

	match := c.checker.FindStringSubmatch(c.args)
	if len(match) != 2 {
		h.sendMessageToChat(c.chatId, "Error")
		return
	}

	userName := match[1]
	existingParticipant, err := h.Storage.FindByLink("@"+userName, c.chatId)
	if err == nil && existingParticipant.Id() != "" {
		h.sendMessageToChat(c.chatId, "User is already in the list of participants")
		return
	}

	if h.Storage.CountByChatId(c.chatId) >= maxParticipants {
		h.sendMessageToChat(c.chatId, fmt.Sprintf("Maximum chat participants: *%v*", maxParticipants))
		return
	}

	participant := h.Storage.Create(
		store.Participant{
			User: store.User{
				UserName: userName,
				Type:     store.UserUnresolved,
			},
			Time:   time.Now(),
			ChatId: c.chatId,
		},
	)

	text := fmt.Sprintf("*Added* %s", store.Escape(participant.Link())) + "\n" +
		h.participantsText(c.chatId)
	h.sendMessageToChat(c.chatId, text)
}

func (h *MessageHandler) addByName(c conversation) {

	existingParticipant, err := h.Storage.FindByName(c.args, c.chatId)
	if err == nil && existingParticipant.Id() != "" {
		h.sendMessageToChat(c.chatId, "User is already in the list of participants")
		return
	}

	if h.Storage.CountByChatId(c.chatId) >= maxParticipants {
		h.sendMessageToChat(c.chatId, fmt.Sprintf("Maximum chat participants: *%v*", maxParticipants))
		return
	}

	participant := h.Storage.Create(
		store.Participant{
			User: store.User{
				UserName: c.args,
				Type:     store.UserGuest,
			},
			Time:   time.Now(),
			ChatId: c.chatId,
		},
	)

	text := fmt.Sprintf("*Added* %s", store.Escape(participant.Link())) + "\n" +
		h.participantsText(c.chatId)
	h.sendMessageToChat(c.chatId, text)
}

func (h *MessageHandler) addByNumber(c conversation) {
	h.sendMessageToChat(c.chatId, "Fail. UserName as an number")
}

func (h *MessageHandler) removeMe(c conversation) {

	participant, err := h.Storage.Find(strconv.Itoa(c.message.From.ID), c.chatId)
	if err != nil {
		participant, err = h.Storage.FindByLink("@"+c.message.From.UserName, c.chatId)
		if err != nil {
			h.sendMessageToChat(c.chatId, "You are not a participant yet")
			return
		}
	}

	h.Storage.Delete(participant)

	text := fmt.Sprintf("*Removed* %s", store.Escape(participant.Link())) + "\n" +
		h.participantsText(c.chatId)
	h.sendMessageToChat(c.chatId, text)
}

func (h *MessageHandler) removeByNumber(c conversation) {

	numberString := string(c.checker.Find([]byte(c.args)))
	number, err := strconv.Atoi(numberString)
	if err != nil {
		h.sendMessageToChat(c.chatId, "Wrong parameter")
		return
	}

	participant, err := h.Storage.FindByNumber(number, c.chatId)
	if err != nil {
		h.sendMessageToChat(c.chatId, store.Escape(err.Error()))
		return
	}

	h.Storage.Delete(participant)

	text := fmt.Sprintf("*Removed* %s", store.Escape(participant.Link())) + "\n" +
		h.participantsText(c.chatId)
	h.sendMessageToChat(c.chatId, text)
}

func (h *MessageHandler) removeByLink(c conversation) {

	linkString := string(c.checker.Find([]byte(c.args)))
	participant, err := h.Storage.FindByLink(linkString, c.chatId)
	if err != nil {
		h.sendMessageToChat(c.chatId, store.Escape(err.Error()))
		return
	}

	h.Storage.Delete(participant)

	text := fmt.Sprintf("*Removed* %s", store.Escape(participant.Link())) + "\n" +
		h.participantsText(c.chatId)
	h.sendMessageToChat(c.chatId, text)
}

func (h *MessageHandler) removeByName(c conversation) {

	participant, err := h.Storage.FindByName(c.args, c.chatId)
	if err != nil {
		h.sendMessageToChat(c.chatId, store.Escape(err.Error()))
		return
	}

	h.Storage.Delete(participant)

	text := fmt.Sprintf("*Removed* %s", store.Escape(participant.Link())) + "\n" +
		h.participantsText(c.chatId)
	h.sendMessageToChat(c.chatId, text)
}

func (h *MessageHandler) reset(c conversation) {
	err := h.Storage.DeleteAll(c.chatId)
	if err != nil {
		h.sendMessageToChat(c.chatId, store.Escape(err.Error()))
		return
	}

	h.sendMessageToChat(c.chatId, "All participants was deleted")
}

func (h *MessageHandler) help(c conversation) {
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
		"_Version: " + h.Version + "_"
	h.sendMessageToChat(c.chatId, text)
}

func (h *MessageHandler) ping(c conversation) {
	h.sendMessageToChat(c.chatId, "Turn to non-participants... *Not implemented*.\nWelcome to https://github.com/taras-by/tbot")
}

func (h *MessageHandler) participantsText(chatId int64) (text string) {
	participants := h.Storage.FindByChatId(chatId)
	if len(participants) == 0 {
		return "No participants"
	}
	text = "Participants:\n"
	for i, p := range participants {
		text = text + fmt.Sprintf(" *%v)* %v\n", i+1, store.Escape(p.Name()))
	}
	return text
}

func (h *MessageHandler) sendMessageToChat(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "markdown"
	_, err := h.Bot.Send(msg)
	if err != nil {
		log.Print(err)
		log.Print(text)
	}
}
