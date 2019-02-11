package main

import (
	"io/ioutil"
	"strings"
)

func param(name, defaultValue string) (value string) {
	ba, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		fatal("could not read /proc/cmdline: ", err)
	}

	prefix := "direktil." + name + "="

	for _, part := range strings.Split(string(ba), " ") {
		if strings.HasPrefix(part, prefix) {
			return strings.TrimSpace(part[len(prefix):])
		}
	}

	return defaultValue
}
