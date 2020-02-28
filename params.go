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

	prefixes := []string{
		"direktil." + name + "=",
		"dkl." + name + "=",
	}

	value = defaultValue

	for _, part := range strings.Split(string(ba), " ") {
		for _, prefix := range prefixes {
			if strings.HasPrefix(part, prefix) {
				value = strings.TrimSpace(part[len(prefix):])
			}
		}
	}

	return
}

func paramBool(name string, defaultValue bool) (value bool) {
	defaultValueS := "false"
	if defaultValue {
		defaultValueS = "true"
	}

	return "true" == param(name, defaultValueS)
}
