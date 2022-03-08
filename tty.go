package main

import (
	"os"

	"golang.org/x/sys/unix"
)

var (
	stdinTTY = &tty{int(os.Stdin.Fd()), nil}
)

type tty struct {
	fd      int
	termios *unix.Termios
}

func (t *tty) EchoOff() {
	termios, err := unix.IoctlGetTermios(t.fd, unix.TCGETS)
	if err != nil {
		return
	}

	t.termios = termios

	newState := *termios
	newState.Lflag &^= unix.ECHO
	newState.Lflag |= unix.ICANON | unix.ISIG
	newState.Iflag |= unix.ICRNL
	unix.IoctlSetTermios(t.fd, unix.TCSETS, &newState)
}

func (t *tty) Restore() {
	if t.termios != nil {
		unix.IoctlSetTermios(t.fd, unix.TCSETS, t.termios)
		t.termios = nil
	}
}
