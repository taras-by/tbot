package main

import (
	"flag"
	"log"
	"os"
)

type command struct {
	fs *flag.FlagSet
	fn func(args []string) error
}

var (
	Opts    = options{}
	Commit  = "Unknown"
	Date    = "Unknown"
	Version = "Unknown"
)

const (
	defaultStorePath = "./bolt.db"
)

func main() {

	commands := map[string]command{
		"run":  runCmd(),
		"show": showCmd(),
	}

	fs := flag.NewFlagSet("tbot", flag.ExitOnError)

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	fs.StringVar(&Opts.TelegramToken, "telegram-token", os.Getenv("TELEGRAM_TOKEN"), "Token for Telegram")
	fs.StringVar(&Opts.StorePath, "store-path", getEnv("STORE_PATH", defaultStorePath), "Path for storage")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatal(err)
	}

	if cmd, ok := commands[args[0]]; !ok {
		log.Fatalf("Unknown command: %s", args[0])
	} else if err := cmd.fn(args[1:]); err != nil {
		log.Fatal(err)
	}
}

func showCmd() command {
	return command{fn: func([]string) error {
		return show()
	}}
}

func runCmd() command {
	return command{fn: func([]string) error {
		a := newApp()
		defer a.storage.Close()
		s := a.makeBotService()
		return s.Run()
	}}
}

func getEnv(key string, value string) string {
	v := os.Getenv(key)
	if v == "" {
		return value
	}
	return v
}
