package main

import (
	nconfig "novit.nc/direktil/pkg/config"
)

type configV1 struct {
	Layers []string          `yaml:"layers"`
	Files  []nconfig.FileDef `yaml:"files"`

	// v2 handles more

	RootUser struct {
		PasswordHash   string   `yaml:"password_hash"`
		AuthorizedKeys []string `yaml:"authorized_keys"`
	} `yaml:"root_user"`

	Mounts []MountDef `yaml:"mounts"`
}

type MountDef struct {
	Dev     string
	Path    string
	Options string
	Type    string
}
