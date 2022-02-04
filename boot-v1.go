package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"
	"novit.nc/direktil/pkg/sysfs"
)

func bootV1() {
	// find and mount /boot
	bootMatch := param("boot", "")
	bootMounted := false
	if bootMatch != "" {
		bootFS := param("boot.fs", "vfat")
		for i := 0; ; i++ {
			devNames := sysfs.DeviceByProperty("block", bootMatch)

			if len(devNames) == 0 {
				if i > 30 {
					fatal("boot partition not found after 30s")
				}
				log.Print("boot partition not found, retrying")
				time.Sleep(1 * time.Second)
				continue
			}

			devFile := filepath.Join("/dev", devNames[0])

			log.Print("boot partition found: ", devFile)

			mount(devFile, "/boot", bootFS, bootMountFlags, "")
			bootMounted = true
			break
		}
	} else {
		log.Print("Assuming /boot is already populated.")
	}

	// load config
	cfgPath := param("config", "/boot/config.yaml")

	cfgBytes, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		fatalf("failed to read %s: %v", cfgPath, err)
	}

	cfg := &config{}
	if err := yaml.Unmarshal(cfgBytes, cfg); err != nil {
		fatal("failed to load config: ", err)
	}

	// mount layers
	if len(cfg.Layers) == 0 {
		fatal("no layers configured!")
	}

	log.Printf("wanted layers: %q", cfg.Layers)

	layersInMemory := paramBool("layers-in-mem", false)

	const layersInMemDir = "/layers-in-mem"
	if layersInMemory {
		mkdir(layersInMemDir, 0700)
		mount("layers-mem", layersInMemDir, "tmpfs", 0, "")
	}

	lowers := make([]string, len(cfg.Layers))
	for i, layer := range cfg.Layers {
		path := layerPath(layer)

		info, err := os.Stat(path)
		if err != nil {
			fatal(err)
		}

		log.Printf("layer %s found (%d bytes)", layer, info.Size())

		if layersInMemory {
			log.Print("  copying to memory...")
			targetPath := filepath.Join(layersInMemDir, layer)
			cp(path, targetPath)
			path = targetPath
		}

		dir := "/layers/" + layer

		lowers[i] = dir

		loopDev := fmt.Sprintf("/dev/loop%d", i)
		losetup(loopDev, path)

		mount(loopDev, dir, "squashfs", layerMountFlags, "")
	}

	// prepare system root
	mount("mem", "/changes", "tmpfs", 0, "")

	mkdir("/changes/workdir", 0755)
	mkdir("/changes/upperdir", 0755)

	mount("overlay", "/system", "overlay", rootMountFlags,
		"lowerdir="+strings.Join(lowers, ":")+",upperdir=/changes/upperdir,workdir=/changes/workdir")

	if bootMounted {
		if layersInMemory {
			if err := syscall.Unmount("/boot", 0); err != nil {
				log.Print("WARNING: failed to unmount /boot: ", err)
				time.Sleep(2 * time.Second)
			}

		} else {
			mount("/boot", "/system/boot", "", syscall.MS_BIND, "")
		}
	}

	// - write configuration
	log.Print("writing /config.yaml")
	if err := ioutil.WriteFile("/system/config.yaml", cfgBytes, 0600); err != nil {
		fatal("failed: ", err)
	}

	// - write files
	for _, fileDef := range cfg.Files {
		log.Print("writing ", fileDef.Path)

		filePath := filepath.Join("/system", fileDef.Path)

		ioutil.WriteFile(filePath, []byte(fileDef.Content), fileDef.Mode)
	}

	// clean zombies
	cleanZombies()

	// switch root
	log.Print("switching root")
	err = syscall.Exec("/sbin/switch_root", []string{"switch_root",
		"-c", "/dev/console", "/system", "/sbin/init"}, os.Environ())
	fatal("switch_root failed: ", err)
}
