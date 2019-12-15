package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
)

type command struct {
	fs *flag.FlagSet
	fn func(args []string) error
}

type Options struct {
	TelegramToken string
	StorePath     string
}

var (
	Opts = Options{}
)

const (
	defaultStorePath = "./bolt.db"
)

func main() {

	commands := map[string]command{
		"server": serverCmd(),
		"show":   showCmd(),
	}

	fs := flag.NewFlagSet("tbot", flag.ExitOnError)

	fs.Parse(os.Args[1:])

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	fs.StringVar(&Opts.TelegramToken, "telegram-token", os.Getenv("TELEGRAM_TOKEN"), "Token for Telegram")
	fs.StringVar(&Opts.StorePath, "store-path", getEnv("STORE_PATH", defaultStorePath), "Path for storage")
	fs.Parse(os.Args[2:])

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

func serverCmd() command {
	return command{fn: func([]string) error {
		return server()
	}}
}

func getEnv(key string, value string) string {
	v := os.Getenv(key)
	if v == "" {
		return value
	}
	return v
}

func printVersion() {
	fmt.Printf("Version: %s\nCommit: %s\nRuntime: %s %s/%s\nDate: %s\n",
		Version,
		Commit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		Date,
	)
}

var (
	Commit  string
	Date    string
	Version string
)
