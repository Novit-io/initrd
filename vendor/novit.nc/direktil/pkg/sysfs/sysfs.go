package sysfs

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

// DeviceByProperty lists the devices where a given property=value filters match.
func DeviceByProperty(class string, filters ...string) []string {
	files, err := filepath.Glob("/sys/class/" + class + "/*/uevent")
	if err != nil {
		log.Print("list devices failed: ", err)
		return nil
	}

	filtered := make([]string, 0)

filesLoop:
	for _, file := range files {
		ba, err := ioutil.ReadFile(file)
		if err != nil {
			log.Print("reading ", file, " failed: ", err)
			continue
		}

		values := strings.Split(strings.TrimSpace(string(ba)), "\n")

		devName := ""
		for _, value := range values {
			if strings.HasPrefix(value, "DEVNAME=") {
				devName = value[len("DEVNAME="):]
			}
		}

		for _, filter := range filters {
			found := false
			for _, value := range values {
				if filter == value {
					found = true
					break
				}
			}

			if !found {
				continue filesLoop
			}
		}

		filtered = append(filtered, devName)
	}

	return filtered
}
