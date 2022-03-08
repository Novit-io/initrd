package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/pkg/term/termios"
	"golang.org/x/term"

	"novit.nc/direktil/initrd/colorio"
	"novit.nc/direktil/initrd/shio"
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

	stdin,
	stdinPipe = newPipe()
	stdout = shio.New()
	stderr = colorio.NewWriter(colorio.Bold, stdout)
)

func newPipe() (io.ReadCloser, io.WriteCloser) {
	return io.Pipe()
}

func main() {
	runtime.LockOSThread()

	// move log to shio
	go io.Copy(os.Stdout, stdout.NewReader())
	log.SetOutput(stderr)

	// copy os.Stdin to my stdin pipe
	go io.Copy(stdinPipe, os.Stdin)

	log.Print("Welcome to ", VERSION)

	// essential mounts
	mount("none", "/proc", "proc", 0, "")
	mount("none", "/sys", "sysfs", 0, "")
	mount("none", "/dev", "devtmpfs", 0, "")
	mount("none", "/dev/pts", "devpts", 0, "gid=5,mode=620")

	// get the "boot version"
	bootVersion = param("version", "current")
	log.Printf("booting system %q", bootVersion)

	os.Setenv("PATH", "/usr/bin:/bin:/usr/sbin:/sbin")

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

var (
	layersDir      = "/boot/current/layers/"
	layersOverride = map[string]string{}
)

func layerPath(name string) string {
	if override, ok := layersOverride[name]; ok {
		return override
	}
	return filepath.Join(layersDir, name+".fs")
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
	log.SetOutput(os.Stderr)
	stdout.Close()
	stdin.Close()
	stdinPipe.Close()

	stdin = nil

mainLoop:
	for {
		termios.Tcdrain(os.Stdin.Fd())
		termios.Tcdrain(os.Stdout.Fd())
		termios.Tcdrain(os.Stderr.Fd())

		fmt.Print("\nr to reboot, o to power off, s to get a shell: ")

		// TODO flush stdin (first char lost here?)
		deadline := time.Now().Add(time.Minute)
		os.Stdin.SetReadDeadline(deadline)

		termios.Tcflush(os.Stdin.Fd(), termios.TCIFLUSH)
		term.MakeRaw(int(os.Stdin.Fd()))

		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			log.Print("failed to read from stdin: ", err)
			time.Sleep(5 * time.Second)
			syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
		}
		fmt.Println()

		switch b[0] {
		case 'o':
			syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
		case 'r':
			syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
		case 's':
			for _, sh := range []string{"bash", "ash", "sh", "busybox"} {
				fullPath, err := exec.LookPath(sh)
				if err != nil {
					continue
				}

				args := make([]string, 0)
				if sh == "busybox" {
					args = append(args, "sh")
				}

				if !localAuth() {
					continue mainLoop
				}

				cmd := exec.Command(fullPath, args...)
				cmd.Env = os.Environ()
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				if err != nil {
					fmt.Println("shell failed:", err)
				}
				continue mainLoop
			}
			log.Print("failed to find a shell!")

		default:
			log.Printf("unknown choice: %q", string(b))
		}
	}
}

func losetup(dev, file string) {
	run("/sbin/losetup", "-r", dev, file)
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
