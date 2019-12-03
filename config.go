package main

import (
	nconfig "novit.nc/direktil/pkg/config"
)

type config struct {
	Layers []string          `yaml:"layers"`
	Files  []nconfig.FileDef `yaml:"files"`
}
