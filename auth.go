package main

import (
	"bytes"
	"errors"
	"log"

	"golang.org/x/crypto/ssh"
	"novit.nc/direktil/initrd/config"
)

var (
	auths []config.Auth
)

func localAuth() bool {
	sec := askSecret("password")

	for _, auth := range auths {
		if auth.Password == "" {
			continue
		}

		if config.CheckPassword(auth.Password, sec) {
			log.Printf("login with auth %q", auth.Name)
			return true
		}
	}

	return false
}

func sshCheckPubkey(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	keyBytes := key.Marshal()

	for _, auth := range auths {
		if auth.SSHKey == "" {
			continue
		}

		allowedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(auth.SSHKey))
		if err != nil {
			log.Printf("SSH pubkey for %q invalid: %v", auth.Name, auth.SSHKey)
			return nil, err
		}

		if bytes.Equal(allowedKey.Marshal(), keyBytes) {
			log.Print("ssh: accepting public key for ", auth.Name)
			return &ssh.Permissions{
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(key),
				},
			}, nil
		}
	}

	return nil, errors.New("no matching public key")
}
