// +build darwin freebsd netbsd openbsd linux

package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

func clearscr() {
	fmt.Printf("%c[2J%c[0;0H", 27, 27)
}

type Term struct {
	s *terminal.State
	t *terminal.Terminal
}

func newTerm() *Term {
	term := new(Term)
	var err error
	term.s, err = terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	term.t = terminal.NewTerminal(os.Stdin, PROMPT)
	return term
}

func (id *Term) Restore() {
	terminal.Restore(0, id.s)
}

func (id *Term) ReadLine() (string, error) {
	return id.t.ReadLine()
}
