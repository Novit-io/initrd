package main

import (
	"log"
	"syscall"
)

func cleanZombies() {
	return // FIXME noop... udhcpc is a daemon staying alive so we never finish

	var wstatus syscall.WaitStatus

	for {
		pid, err := syscall.Wait4(-1, &wstatus, 0, nil)
		switch err {
		case nil:
			log.Printf("collected PID %v", pid)

		case syscall.ECHILD:
			return

		default:
			log.Printf("unknown error: %v", err)
		}
	}
}
