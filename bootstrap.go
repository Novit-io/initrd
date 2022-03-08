package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"novit.nc/direktil/initrd/config"
)

func bootstrap(cfg *config.Config) {
	if cfg.Bootstrap.Dev == "" {
		fatalf("bootstrap device not defined!")
	}

	const bsDir = "/bootstrap"
	os.MkdirAll(bsDir, 0700)

	run("mount", cfg.Bootstrap.Dev, bsDir)

	baseDir := filepath.Join(bsDir, bootVersion)
	sysCfgPath := filepath.Join(baseDir, "config.yaml")

	if _, err := os.Stat(sysCfgPath); os.IsNotExist(err) {
		log.Printf("bootstrap %q does not exist", bootVersion)

		seed := cfg.Bootstrap.Seed
		if seed == "" {
			fatalf("boostrap seed not defined, admin required")
		}

		log.Printf("seeding bootstrap from %s", seed)

		// TODO
	}

	layersDir = baseDir
	layersOverride["modules"] = "/modules.sqfs"
	sysCfg := applyConfig(sysCfgPath, false)

	// mounts are v2 only
	for _, mount := range sysCfg.Mounts {
		log.Print("mount ", mount.Dev, " to system's ", mount.Path)

		path := filepath.Join("/system", mount.Path)

		os.MkdirAll(path, 0755)

		args := []string{mount.Dev, path}
		if mount.Type != "" {
			args = append(args, "-t", mount.Type)
		}
		if mount.Options != "" {
			args = append(args, "-o", mount.Options)
		}

		run("mount", args...)
	}

	// setup root user
	if ph := sysCfg.RootUser.PasswordHash; ph != "" {
		log.Print("setting root's password")
		setUserPass("root", ph)
	}
	if ak := sysCfg.RootUser.AuthorizedKeys; len(ak) != 0 {
		log.Print("setting root's authorized keys")
		setAuthorizedKeys(ak)
	}
}

func setUserPass(user, passwordHash string) {
	const fpath = "/system/etc/shadow"

	ba, err := ioutil.ReadFile(fpath)
	if err != nil {
		fatalf("failed to read shadow: %v", err)
	}

	lines := bytes.Split(ba, []byte{'\n'})

	buf := new(bytes.Buffer)
	for _, line := range lines {
		line := string(line)
		p := strings.Split(line, ":")
		if len(p) < 2 || p[0] != user {
			buf.WriteString(line)
			continue
		}

		p[1] = passwordHash
		line = strings.Join(p, ":")

		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	err = ioutil.WriteFile(fpath, buf.Bytes(), 0600)
	if err != nil {
		fatalf("failed to write shadow: %v", err)
	}
}

func setAuthorizedKeys(ak []string) {
	buf := new(bytes.Buffer)
	for _, k := range ak {
		buf.WriteString(k)
		buf.WriteByte('\n')
	}

	const sshDir = "/system/root/.ssh"
	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		fatalf("failed to create %s: %v", sshDir, err)
	}

	err = ioutil.WriteFile(filepath.Join(sshDir, "authorized_keys"), buf.Bytes(), 0600)
	if err != nil {
		fatalf("failed to write authorized keys: %v", err)
	}
}
