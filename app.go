package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/taras-by/tbot/store"
	tlg "github.com/taras-by/tbot/telegram"
	"log"
	"runtime"
)

type options struct {
	TelegramToken string
	StorePath     string
}

type app struct {
	options options
	commit  string
	date    string
	version string
	storage *store.Storage
}

func newApp() (a *app) {

	a = &app{
		options: Opts,
		commit:  Commit,
		date:    Date,
		version: Version,
	}
	a.printVersion()

	var err error

	a.storage, err = store.NewStorage(a.options.StorePath)
	if err != nil {
		log.Printf("storage creating error. Path: %s", a.options.StorePath)
		log.Panic(err.Error())
	}

	return a
}

func (a *app) makeBotService() (s *tlg.BotService) {

	bot, err := tgbotapi.NewBotAPI(a.options.TelegramToken)
	if err != nil {
		log.Printf("Telegram connection Error. Token: %s", a.options.TelegramToken)
		log.Panic(err.Error())
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	handler := &tlg.MessageHandler{
		Bot:     bot,
		Storage: a.storage,
		Version: a.version,
	}

	service := tlg.BotService{
		Bot:     bot,
		Handler: handler,
	}
	service.Init()
	return &service
}

func (a *app) printVersion() {
	fmt.Printf("Version: %s\nCommit: %s\nRuntime: %s %s/%s\nDate: %s\n",
		a.version,
		a.commit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		a.date,
	)
}

func (a *app) Close() () {
	a.storage.Close()
}
