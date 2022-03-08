package main

import (
	"encoding/json"
	"os/exec"
)

func runJSON(v interface{}, cmd string, args ...string) (err error) {
	c := exec.Command(cmd, args...)
	c.Stderr = stderr
	ba, err := c.Output()
	if err != nil {
		return
	}

	err = json.Unmarshal(ba, v)
	return
}
