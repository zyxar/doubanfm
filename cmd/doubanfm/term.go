//+build darwin freebsd netbsd openbsd linux

package main

import (
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type Term struct {
	s *terminal.State
	t *terminal.Terminal
}

func newTerm(prompt string) *Term {
	term := new(Term)
	var err error
	term.s, err = terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	term.t = terminal.NewTerminal(os.Stdin, prompt)
	return term
}

func (this *Term) Restore() {
	terminal.Restore(0, this.s)
}

func (this *Term) ReadLine() (string, error) {
	return this.t.ReadLine()
}

func (this *Term) ReadPassword(prompt string) (string, error) {
	return this.t.ReadPassword(prompt)
}
