package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"

	"novit.nc/direktil/initrd/config"
)

func startSSH(cfg *config.Config) {
	sshConfig := &ssh.ServerConfig{
		PublicKeyCallback: sshCheckPubkey,
	}

	pkBytes, err := ioutil.ReadFile("/id_rsa") // TODO configurable
	if err != nil {
		fatalf("ssh: failed to load private key: %v", err)
	}

	pk, err := ssh.ParsePrivateKey(pkBytes)
	if err != nil {
		fatalf("ssh: failed to parse private key: %v", err)
	}

	sshConfig.AddHostKey(pk)

	sshBind := ":22" // TODO configurable
	listener, err := net.Listen("tcp", sshBind)
	if err != nil {
		fatalf("ssh: failed to listen on %s: %v", sshBind, err)
	}

	log.Print("SSH server listening on ", sshBind)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Print("ssh: accept conn failed: ", err)
				continue
			}

			go sshHandleConn(conn, sshConfig)
		}
	}()
}

func sshHandleConn(conn net.Conn, sshConfig *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, sshConfig)

	if err != nil {
		log.Print("ssh: handshake failed: ", err)
		return
	}

	remoteAddr := sshConn.User() + "@" + sshConn.RemoteAddr().String()
	log.Print("ssh: new connection from ", remoteAddr)

	go sshHandleReqs(reqs)
	go sshHandleChannels(remoteAddr, chans)
}

func sshHandleReqs(reqs <-chan *ssh.Request) {
	for req := range reqs {
		switch req.Type {
		case "keepalive@openssh.com":
			req.Reply(true, nil)

		default:
			log.Printf("ssh: discarding req: %+v", req)
			req.Reply(false, nil)
		}
	}
}

func sshHandleChannels(remoteAddr string, chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Print("ssh: failed to accept channel: ", err)
			continue
		}

		go sshHandleChannel(remoteAddr, channel, requests)
	}
}

func sshHandleChannel(remoteAddr string, channel ssh.Channel, requests <-chan *ssh.Request) {
	var (
		ptyF, ttyF *os.File
		termEnv    string
	)

	defer func() {
		if ptyF != nil {
			ptyF.Close()
		}
		if ttyF != nil {
			ttyF.Close()
		}
	}()

	var once sync.Once
	close := func() {
		channel.Close()
	}

	for req := range requests {
		switch req.Type {
		case "exec":
			command := string(req.Payload[4 : req.Payload[3]+4])
			switch command {
			case "init":
				go func() {
					io.Copy(channel, stdout.NewReader())
					once.Do(close)
				}()
				go func() {
					io.Copy(stdinPipe, channel)
					once.Do(close)
				}()

				req.Reply(true, nil)

			default:
				req.Reply(false, nil)
			}

		case "shell":
			cmd := exec.Command("/bin/ash")
			cmd.Env = []string{"TERM=" + termEnv}
			cmd.Stdin = ttyF
			cmd.Stdout = ttyF
			cmd.Stderr = ttyF
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setctty:   true,
				Setsid:    true,
				Pdeathsig: syscall.SIGKILL,
			}

			cmd.Start()

			go func() {
				cmd.Wait()
				ptyF.Close()
				ptyF = nil
				ttyF.Close()
				ttyF = nil
			}()

			go func() {
				io.Copy(channel, ptyF)
				once.Do(close)
			}()
			go func() {
				io.Copy(ptyF, channel)
				once.Do(close)
			}()

			req.Reply(true, nil)

		case "pty-req":
			if ptyF != nil || ttyF != nil {
				req.Reply(false, nil)
				continue
			}

			var err error
			ptyF, ttyF, err = pty.Open()
			if err != nil {
				log.Print("PTY err: ", err)
				req.Reply(false, nil)
				continue
			}

			termLen := req.Payload[3]
			termEnv = string(req.Payload[4 : termLen+4])
			w, h := sshParseDims(req.Payload[termLen+4:])
			sshSetWinsize(ptyF.Fd(), w, h)

			req.Reply(true, nil)

		case "window-change":
			w, h := sshParseDims(req.Payload)
			sshSetWinsize(ptyF.Fd(), w, h)
			// no response

		default:
			req.Reply(false, nil)
		}
	}
}

func sshParseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// SetWinsize sets the size of the given pty.
func sshSetWinsize(fd uintptr, w, h uint32) {
	// Winsize stores the Height and Width of a terminal.
	type Winsize struct {
		Height uint16
		Width  uint16
		x      uint16 // unused
		y      uint16 // unused
	}

	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
