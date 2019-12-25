package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BotService struct {
	Bot     *tgbotapi.BotAPI
	Handler *MessageHandler
}

func (s *BotService) Init() {
	h := s.Handler
	h.routes = []route{
		{`add`, ``, h.addMe},
		{`add`, `^@(\S+)$`, h.addByLink},
		{`add`, `^\d+$`, h.addByNumber},
		{`add`, `^.+$`, h.addByName},
		{`rm`, ``, h.removeMe},
		{`rm`, `^@(\S+)$`, h.removeByLink},
		{`rm`, `^\d+$`, h.removeByNumber},
		{`rm`, `^.+$`, h.removeByName},
		{`list`, ``, h.list},
		{`ping`, ``, h.ping},
		{`reset`, ``, h.reset},
		{`start`, ``, h.help},
		{`help`, ``, h.help},
	}
}

func (s *BotService) Run() error {

	updates, err := s.chatUpdates()
	if err != nil {
		return err
	}

	for update := range updates {
		s.Handler.handle(update.Message)
	}
	return nil
}

func (s *BotService) chatUpdates() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := s.Bot.GetUpdatesChan(u)
	return updates, err
}
