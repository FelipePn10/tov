package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/FelipePn10/tov/internal/repo"
)

type command struct {
	name    string // "init" ,"commit" etc
	summary string
	run     func(args []string) error // recebe os argumentos que vieram depois do nome do comando
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
		fmt.Fprintf(os.Stderr, "tov: '%s' não é um comando. Veja 'tov help'. \n", name)
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

func cmdInit(args []string) error {
	fs := newFlags("init", "[--bare] [diretório]")
	bare := fs.Bool("bare", false, "cria um repositório sem árvore de trabalho")
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}
	r, err := repo.Init(dir, *bare)
	if err != nil {
		return err
	}
	fmt.Printf("Repositório vazio criado em %s\n", r.GitDir)
	return nil
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
