package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/term"
)

const (
	// VERSION is the current version of init
	VERSION = "Direktil init v2.0"

	rootMountFlags  = 0
	bootMountFlags  = syscall.MS_NOEXEC | syscall.MS_NODEV | syscall.MS_NOSUID | syscall.MS_RDONLY
	layerMountFlags = syscall.MS_RDONLY
)

var (
	bootVersion string
)

func main() {
	runtime.LockOSThread()

	log.Print("Welcome to ", VERSION)

	// essential mounts
	mount("none", "/proc", "proc", 0, "")
	mount("none", "/sys", "sysfs", 0, "")
	mount("none", "/dev", "devtmpfs", 0, "")

	// get the "boot version"
	bootVersion = param("version", "current")
	log.Printf("booting system %q", bootVersion)

	_, err := os.Stat("/config.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			bootV1()
			return
		}
		fatal("stat failed: ", err)
	}

	bootV2()
}

func layerPath(name string) string {
	return fmt.Sprintf("/boot/%s/layers/%s.fs", bootVersion, name)
}

func fatal(v ...interface{}) {
	log.Print("*** FATAL ***")
	log.Print(v...)
	die()
}

func fatalf(pattern string, v ...interface{}) {
	log.Print("*** FATAL ***")
	log.Printf(pattern, v...)
	die()
}

func die() {
	fmt.Println("\nwill reboot in 1 minute; press r to reboot now, o to power off, s to get a shell")

	deadline := time.Now().Add(time.Minute)

	term.MakeRaw(int(os.Stdin.Fd())) // disable line buffering
	os.Stdin.SetReadDeadline(deadline)

	b := []byte{0}
	for {
		_, err := os.Stdin.Read(b)
		if err != nil {
			break
		}

		fmt.Println(string(b))

		switch b[0] {
		case 'o':
			syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
		case 'r':
			syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
		case 's':
			err = syscall.Exec("/bin/ash", []string{"/bin/ash"}, os.Environ())
			if err != nil {
				fmt.Println("failed to start the shell:", err)
			}
		}
	}

	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func losetup(dev, file string) {
	run("/sbin/losetup", dev, file)
}

func run(cmd string, args ...string) {
	if output, err := exec.Command(cmd, args...).CombinedOutput(); err != nil {
		fatalf("command %s %q failed: %v\n%s", cmd, args, err, string(output))
	}
}

func mkdir(dir string, mode os.FileMode) {
	if err := os.MkdirAll(dir, mode); err != nil {
		fatalf("mkdir %q failed: %v", dir, err)
	}
}

func mount(source, target, fstype string, flags uintptr, data string) {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		mkdir(target, 0755)
	}

	if err := syscall.Mount(source, target, fstype, flags, data); err != nil {
		fatalf("mount %q %q -t %q -o %q failed: %v", source, target, fstype, data, err)
	}
	log.Printf("mounted %q", target)
}

func cp(srcPath, dstPath string) {
	var err error
	defer func() {
		if err != nil {
			fatalf("cp %s %s failed: %v", srcPath, dstPath, err)
		}
	}()

	src, err := os.Open(srcPath)
	if err != nil {
		return
	}

	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return
	}

	defer dst.Close()

	_, err = io.Copy(dst, src)
}
