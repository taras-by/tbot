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

func main() {

	commands := map[string]command{
		"server": serverCmd(),
		"show": showCmd(),
	}

	fs := flag.NewFlagSet("tbot", flag.ExitOnError)

	fs.Parse(os.Args[1:])

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(1)
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

func serverCmd() command {
	return command{fn: func([]string) error {
		return server()
	}}
}
