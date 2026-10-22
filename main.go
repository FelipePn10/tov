package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
)

type command struct {
	name    string
	summary string
	run     func(args []string) error
}

var commands []command

func register(name, summary string, run func(args []string) error) {
	commands = append(commands, command{name, summary, run})
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		usage()
		return
	}
	name := os.Args[1]
	i := slices.IndexFunc(commands, func(c command) bool { return c.name == name })
	if i < 0 {
		fmt.Fprintf(os.Stderr, "tov: '%s' não é um comando. Veja 'kit help'. \n", name)
		os.Exit(1)
	}
	if err := commands[i].run(os.Args[2:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		var ex exitError
		if errors.As(err, &ex) {
			os.Exit(int(ex))
		}
		fmt.Fprintf(os.Stderr, "kit %s: %v\n", name, err)
		os.Exit(1)
	}
}

type exitError int

func usage() {
	fmt.Println("uso: tov <comando> [argumentos]")
	fmt.Println()
	slices.SortFunc(commands, func(a, b command) int {
		return strings.Compare(a.name, b.name)
	})
	for _, c := range commands {
		fmt.Printf("  %-14s %s\n", c.name, c.summary)
	}
}

func newFlags(name, args string) *flag.FlagSet {
	fs := flag.NewFlagSet("tov "+name, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "uso: tov %s %s\n", name, args)
		fs.PrintDefaults()
	}
	return fs
}
