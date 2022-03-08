package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

func askSecret(prompt string) []byte {
	stdinTTY.EchoOff()

	var (
		in  io.Reader = stdin
		out io.Writer = stdout
	)

	if stdin == nil {
		in = os.Stdin
		out = os.Stdout
	}

	out.Write([]byte(prompt + ": "))

	if stdin != nil {
		stdout.HideInput()
	}

	s, err := bufio.NewReader(in).ReadBytes('\n')

	if stdin != nil {
		stdout.ShowInput()
	}

	stdinTTY.Restore()

	if err != nil {
		fatalf("failed to read from stdin: %v", err)
	}

	s = bytes.TrimRight(s, "\r\n")
	return s
}
