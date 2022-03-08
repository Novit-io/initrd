package main

import (
	"log"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"

	"novit.nc/direktil/initrd/config"
)

func bootV2() {
	log.Print("-- boot v2 --")

	cfg := &config.Config{}

	{
		f, err := os.Open("/config.yaml")
		if err != nil {
			fatal("failed to open /config.yaml: ", err)
		}

		err = yaml.NewDecoder(f).Decode(cfg)
		f.Close()

		if err != nil {
			fatal("failed to parse /config.yaml: ", err)
		}
	}

	log.Print("config loaded")
	log.Printf("anti-phishing code: %q", cfg.AntiPhishingCode)

	auths = cfg.Auths

	// mount kernel modules
	if cfg.Modules != "" {
		log.Print("mount modules from ", cfg.Modules)

		err := os.MkdirAll("/modules", 0755)
		if err != nil {
			fatal("failed to create /modules: ", err)
		}

		run("mount", cfg.Modules, "/modules")
		loopOffset++

		err = os.Symlink("/modules/lib/modules", "/lib/modules")
		if err != nil {
			fatal("failed to symlink modules: ", err)
		}
	}

	// load basic modules
	run("modprobe", "unix")

	// devices init
	err := exec.Command("udevd").Start()
	if err != nil {
		fatal("failed to start udevd: ", err)
	}

	log.Print("udevadm triggers")
	run("udevadm", "trigger", "-c", "add", "-t", "devices")
	run("udevadm", "trigger", "-c", "add", "-t", "subsystems")

	log.Print("udevadm settle")
	run("udevadm", "settle")

	// networks
	setupNetworks(cfg)

	// Wireguard VPN
	// TODO startVPN()

	// SSH service
	startSSH(cfg)

	// LVM
	setupLVM(cfg)

	// bootstrap the system
	bootstrap(cfg)

	// finalize
	finalizeBoot()
}
