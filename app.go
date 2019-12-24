package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	srv "github.com/taras-by/tbot/server"
	"github.com/taras-by/tbot/store"
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
	bot     *tgbotapi.BotAPI
	storage *store.Storage
}

func newApp() (a *app) {

	a = &app{
		options: Opts,
		commit:  Commit,
		date:    Date,
		version: Version,
	}

	var err error

	a.bot, err = tgbotapi.NewBotAPI(a.options.TelegramToken)
	if err != nil {
		log.Printf("Telegram connection Error. Token: %s", a.options.TelegramToken)
		log.Panic(err.Error())
	}

	a.storage, err = store.NewStorage(a.options.StorePath)
	if err != nil {
		log.Printf("storage creating error. Path: %s", a.options.StorePath)
		log.Panic(err.Error())
	}

	return a
}

func (a *app) newServer() (s *srv.Server) {
	return &srv.Server{
		Bot:     a.bot,
		Storage: a.storage,
		Version: a.version,
	}
}

func (a app) printVersion() {
	fmt.Printf("Version: %s\nCommit: %s\nRuntime: %s %s/%s\nDate: %s\n",
		a.version,
		a.commit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		a.date,
	)
}
