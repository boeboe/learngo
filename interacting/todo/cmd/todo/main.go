package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/boeboe/learngo/interacting/todo"
)

var todoFileName = ".todo.json"

func main() {
	if os.Getenv("TODO_FILENAME") != "" {
		todoFileName = os.Getenv("TODO_FILENAME")
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s tool. Developed by boeboe\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Copyright 2022\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage information\n")
		flag.PrintDefaults()
	}

	add := flag.Bool("add", false, "add task to the list")
	list := flag.Bool("list", false, "list all tasks")
	complete := flag.Int("complete", 0, "todo item to complete")
	delete := flag.Int("delete", 0, "todo item to delete")
	flag.Parse()

	l := &todo.List{}

	if err := l.Get(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case *add:
		t, err := getTask(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		l.Add(t)
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *list:
		fmt.Print(l)
	case *complete > 0:
		if err := l.Complete(*complete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *delete > 0:
		if err := l.Delete(*delete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "invalid option")
		os.Exit(1)
	}
}

func getTask(r io.Reader, args ...string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	s := bufio.NewScanner(r)
	s.Scan()

	if err := s.Err(); err != nil {
		return "", err
	}
	if len(s.Text()) == 0 {
		return "", fmt.Errorf("task cannot be blank")
	}
	return s.Text(), nil
}
