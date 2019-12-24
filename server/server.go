package server

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

type route struct {
	botCommand    string
	argExpression string
	command       func(c conversation)
}

type Server struct {
	Bot     *tgbotapi.BotAPI
	Storage *store.Storage
	Version string
}

type conversation struct {
	chatId  int64
	args    string
	checker *regexp.Regexp
	message *tgbotapi.Message
}

func (s *Server) routes() []route {
	return []route{
		{`add`, ``, s.addMe},
		{`add`, `^@(\S+)$`, s.addByLink},
		{`add`, `^\d+$`, s.addByNumber},
		{`add`, `^.+$`, s.addByName},
		{`rm`, ``, s.removeMe},
		{`rm`, `^@(\S+)$`, s.removeByLink},
		{`rm`, `^\d+$`, s.removeByNumber},
		{`rm`, `^.+$`, s.removeByName},
		{`list`, ``, s.list},
		{`ping`, ``, s.ping},
		{`reset`, ``, s.reset},
		{`start`, ``, s.help},
		{`help`, ``, s.help},
	}
}

func (s *Server) Run() error {

	defer s.Storage.Close()

	log.Printf("Authorized on account %s", s.Bot.Self.UserName)
	routes := s.routes()
	updates, err := s.chatUpdates()
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("Message: [%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() == false { // ignore any non-command Updates
			continue
		}

		args := strings.TrimSpace(update.Message.CommandArguments())
		cmd := update.Message.Command()
		chatId := update.Message.Chat.ID

		if len([]rune(args)) > maxLengthStringArgument {
			s.sendMessageToChat(chatId, store.Escape("Parameter too long"))
			continue
		}

		commandIsOk := false
		for _, route := range routes {

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
					message: update.Message,
				}

				route.command(c)
				break
			}
		}
		if commandIsOk == false {
			s.sendMessageToChat(chatId, store.Escape("Wrong command"))
		}
	}
	return nil
}

func (s *Server) chatUpdates() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := s.Bot.GetUpdatesChan(u)
	return updates, err
}

func (s *Server) list(c conversation) {
	text := s.participantsText(c.chatId)
	s.sendMessageToChat(c.chatId, text)
}

func (s *Server) addMe(c conversation) {

	creationTime := time.Now()
	existingParticipant, err := s.Storage.FindByLink("@"+c.message.From.UserName, c.chatId)
	if err == nil {
		if existingParticipant.IsUnresolved() == true {
			creationTime = existingParticipant.Time
			s.Storage.Delete(existingParticipant)
			//sendMessageToChat(chatId, "Update unresolved")
		} else {
			s.sendMessageToChat(c.chatId, "You are already a participant")
			return
		}
	}

	participant := s.Storage.Create(
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

	s.sendMessageToChat(c.chatId, fmt.Sprintf("Added %s", store.Escape(participant.Link())))
	text := s.participantsText(c.chatId)
	s.sendMessageToChat(c.chatId, text)
}

func (s *Server) addByLink(c conversation) {

	match := c.checker.FindStringSubmatch(c.args)
	if len(match) != 2 {
		s.sendMessageToChat(c.chatId, "Error")
		return
	}

	userName := match[1]
	existingParticipant, err := s.Storage.FindByLink("@"+userName, c.chatId)
	if err == nil && existingParticipant.Id() != "" {
		s.sendMessageToChat(c.chatId, "User is already in the list of participants")
		return
	}

	participant := s.Storage.Create(
		store.Participant{
			User: store.User{
				UserName: userName,
				Type:     store.UserUnresolved,
			},
			Time:   time.Now(),
			ChatId: c.chatId,
		},
	)

	s.sendMessageToChat(c.chatId, fmt.Sprintf("Added %s", store.Escape(participant.Link())))
	text := s.participantsText(c.chatId)
	s.sendMessageToChat(c.chatId, text)
}

func (s *Server) addByName(c conversation) {

	participant := s.Storage.Create(
		store.Participant{
			User: store.User{
				UserName: c.args,
				Type:     store.UserGuest,
			},
			Time:   time.Now(),
			ChatId: c.chatId,
		},
	)

	s.sendMessageToChat(c.chatId, fmt.Sprintf("Added %s", store.Escape(participant.Link())))
	text := s.participantsText(c.chatId)
	s.sendMessageToChat(c.chatId, text)
}

func (s *Server) addByNumber(c conversation) {
	s.sendMessageToChat(c.chatId, "Fail. UserName as an number")
}

func (s *Server) removeMe(c conversation) {

	var participant store.Participant
	var err error

	participant, err = s.Storage.Find(strconv.Itoa(c.message.From.ID), c.chatId)
	if err != nil {
		participant, err = s.Storage.FindByLink("@"+c.message.From.UserName, c.chatId)
		if err != nil {
			s.sendMessageToChat(c.chatId, "You are not a participant yet")
			return
		}
	}

	s.Storage.Delete(participant)
	s.sendMessageToChat(c.chatId, fmt.Sprintf("Removed %s", store.Escape(participant.Link())))

	text := s.participantsText(c.chatId)
	s.sendMessageToChat(c.chatId, text)
}

func (s *Server) removeByNumber(c conversation) {

	var participant store.Participant
	var err error

	numberString := string(c.checker.Find([]byte(c.args)))
	number, err := strconv.Atoi(numberString)
	if err != nil {
		s.sendMessageToChat(c.chatId, "Wrong parameter")
		return
	}

	participant, err = s.Storage.FindByNumber(number, c.chatId)
	if err != nil {
		s.sendMessageToChat(c.chatId, store.Escape(err.Error()))
		return
	}

	s.Storage.Delete(participant)
	s.sendMessageToChat(c.chatId, fmt.Sprintf("Removed %s", store.Escape(participant.Link())))

	text := s.participantsText(c.chatId)
	s.sendMessageToChat(c.chatId, text)
}

func (s *Server) removeByLink(c conversation) {

	var participant store.Participant
	var err error

	linkString := string(c.checker.Find([]byte(c.args)))
	participant, err = s.Storage.FindByLink(linkString, c.chatId)
	if err != nil {
		s.sendMessageToChat(c.chatId, store.Escape(err.Error()))
		return
	}

	s.Storage.Delete(participant)
	s.sendMessageToChat(c.chatId, fmt.Sprintf("Removed %s", store.Escape(participant.Link())))

	text := s.participantsText(c.chatId)
	s.sendMessageToChat(c.chatId, text)
}

func (s *Server) removeByName(c conversation) {

	var participant store.Participant
	var err error

	participant, err = s.Storage.FindByName(c.args, c.chatId)
	if err != nil {
		s.sendMessageToChat(c.chatId, store.Escape(err.Error()))
		return
	}

	s.Storage.Delete(participant)
	s.sendMessageToChat(c.chatId, fmt.Sprintf("Removed %s", store.Escape(participant.Link())))

	text := s.participantsText(c.chatId)
	s.sendMessageToChat(c.chatId, text)
}

func (s *Server) reset(c conversation) {
	err := s.Storage.DeleteAll(c.chatId)
	if err != nil {
		s.sendMessageToChat(c.chatId, store.Escape(err.Error()))
		return
	}

	s.sendMessageToChat(c.chatId, "All participants was deleted")
}

func (s *Server) help(c conversation) {
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
		"_Version: " + s.Version + "_"
	s.sendMessageToChat(c.chatId, text)
}

func (s *Server) ping(c conversation) {
	s.sendMessageToChat(c.chatId, "Turn to non-participants... *Not implemented*.\nWelcome to https://github.com/taras-by/tbot")
}

func (s *Server) participantsText(chatId int64) (text string) {
	participants := s.Storage.FindByChatId(chatId)
	if len(participants) == 0 {
		return "No participants"
	}
	text = "List of participants:\n"
	for i, p := range participants {
		text = text + fmt.Sprintf(" *%v)* %v\n", i+1, store.Escape(p.Name()))
	}
	return text
}

func (s *Server) sendMessageToChat(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "markdown"
	_, err := s.Bot.Send(msg)
	if err != nil {
		log.Print(err)
		log.Print(text)
	}
}
