package main

import (
	"log"
	"net"
	"os/exec"
	"strings"

	"novit.nc/direktil/initrd/config"
)

func setupNetworks(cfg *config.Config) {
	if len(cfg.Networks) == 0 {
		log.Print("no networks configured.")
		return
	}

	ifNames := make([]string, 0)
	{
		ifaces, err := net.Interfaces()
		if err != nil {
			fatal("failed")
		}
		for _, iface := range ifaces {
			ifNames = append(ifNames, iface.Name)
		}
	}

	assigned := map[string]bool{}

	for _, network := range cfg.Networks {
		log.Print("setting up network ", network.Name)

		// compute available names
		if len(assigned) != 0 {
			newNames := make([]string, 0, len(ifNames))
			for _, n := range ifNames {
				if assigned[n] {
					continue
				}
				newNames = append(newNames, n)
			}
			ifNames = newNames
		}

		// assign envvars
		envvars := make([]string, 0, 1+len(network.Interfaces))
		envvars = append(envvars, "PATH=/bin:/sbin:/usr/bin:/usr/sbin")

		for _, match := range network.Interfaces {
			envvar := new(strings.Builder)
			envvar.WriteString(match.Var)
			envvar.WriteByte('=')

			for i, m := range regexpSelectN(match.N, match.Regexps, ifNames) {
				if i != 0 {
					envvar.WriteByte(' ')
				}
				envvar.WriteString(m)
				assigned[m] = true
			}

			log.Print("- ", envvar)
			envvars = append(envvars, envvar.String())
		}

		cmd := exec.Command("/bin/sh", "-c", network.Script)
		cmd.Env = envvars
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			fatal("failed to setup network: ", err)
		}
	}
}
