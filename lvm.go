package main

import (
	"bytes"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"novit.nc/direktil/initrd/config"
	"novit.nc/direktil/initrd/lvm"
)

func setupLVM(cfg *config.Config) {
	if len(cfg.LVM) == 0 {
		log.Print("no LVM VG configured.")
		return
	}

	run("pvscan")
	run("vgscan", "--mknodes")

	for _, vg := range cfg.LVM {
		setupVG(vg)
	}

	for _, vg := range cfg.LVM {
		setupLVs(vg)
	}

	run("vgchange", "--sysinit", "-a", "ly")

	for _, vg := range cfg.LVM {
		setupCrypt(vg)
	}

	for _, vg := range cfg.LVM {
		setupFS(vg)
	}
}

func setupVG(vg config.LvmVG) {
	pvs := lvm.PVSReport{}
	err := runJSON(&pvs, "pvs", "--reportformat", "json")
	if err != nil {
		fatalf("failed to list LVM PVs: %v", err)
	}

	vgExists := false
	devNeeded := vg.PVs.N
	for _, pv := range pvs.PVs() {
		if pv.VGName == vg.VG {
			vgExists = true
			devNeeded--
		}
	}

	if devNeeded <= 0 {
		log.Print("LVM VG ", vg.VG, " has all its devices")
		return
	}

	if vgExists {
		log.Printf("LVM VG %s misses %d devices", vg.VG, devNeeded)
	} else {
		log.Printf("LVM VG %s does not exists, creating", vg.VG)
	}

	devNames := make([]string, 0)
	err = filepath.Walk("/dev", func(n string, fi fs.FileInfo, err error) error {
		if fi.Mode().Type() == os.ModeDevice {
			devNames = append(devNames, n)
		}
		return err
	})
	if err != nil {
		fatalf("failed to walk /dev: %v", err)
	}

	devNames = filter(devNames, func(v string) bool {
		for _, pv := range pvs.PVs() {
			if v == pv.Name {
				return false
			}
		}
		return true
	})

	m := regexpSelectN(vg.PVs.N, vg.PVs.Regexps, devNames)
	if len(m) == 0 {
		log.Printf("no devices match the regexps %v", vg.PVs.Regexps)
		for _, d := range devNames {
			log.Print("- ", d)
		}

		fatalf("failed to setup VG %s", vg.VG)
	}

	if vgExists {
		log.Print("- extending vg to ", m)
		run("vgextend", append([]string{vg.VG}, m...)...)
		devNeeded -= len(m)
	} else {
		log.Print("- creating vg with devices ", m)
		run("vgcreate", append([]string{vg.VG}, m...)...)
		devNeeded -= len(m)
	}

	if devNeeded > 0 {
		fatalf("VG %s does not have enough devices (%d missing)", vg.VG, devNeeded)
	}
}

func setupLVs(vg config.LvmVG) {
	lvsRep := lvm.LVSReport{}
	err := runJSON(&lvsRep, "lvs", "--reportformat", "json")
	if err != nil {
		fatalf("lvs failed: %v", err)
	}

	lvs := lvsRep.LVs()

	defaults := vg.Defaults

	for _, lv := range vg.LVs {
		lvKey := vg.VG + "/" + lv.Name

		if contains(lvs, func(v lvm.LV) bool {
			return v.VGName == vg.VG && v.Name == lv.Name
		}) {
			log.Printf("LV %s exists", lvKey)
			continue
		}

		log.Printf("creating LV %s", lvKey)

		if lv.Raid == nil {
			lv.Raid = defaults.Raid
		}

		args := make([]string, 0)

		if lv.Name == "" {
			fatalf("LV has no name")
		}
		args = append(args, vg.VG, "--name", lv.Name)

		if lv.Size != "" && lv.Extents != "" {
			fatalf("LV has both size and extents defined!")
		} else if lv.Size == "" && lv.Extents == "" {
			fatalf("LV does not have size or extents defined!")
		} else if lv.Size != "" {
			args = append(args, "-L", lv.Size)
		} else /* if lv.Extents != "" */ {
			args = append(args, "-l", lv.Extents)
		}

		if raid := lv.Raid; raid != nil {
			if raid.Mirrors != 0 {
				args = append(args, "--mirrors", strconv.Itoa(raid.Mirrors))
			}
			if raid.Stripes != 0 {
				args = append(args, "--stripes", strconv.Itoa(raid.Stripes))
			}
		}

		log.Print("lvcreate args: ", args)
		run("lvcreate", args...)

		dev := "/dev/" + vg.VG + "/" + lv.Name
		zeroDevStart(dev)
	}
}

func zeroDevStart(dev string) {
	f, err := os.OpenFile(dev, os.O_WRONLY, 0600)
	if err != nil {
		fatalf("failed to open %s: %v", dev, err)
	}

	defer f.Close()

	_, err = f.Write(make([]byte, 8192))
	if err != nil {
		fatalf("failed to zero the beginning of %s: %v", dev, err)
	}
}

func setupCrypt(vg config.LvmVG) {
	cryptDevs := map[string]bool{}

	var password []byte
	passwordVerified := false

	for _, lv := range vg.LVs {
		if lv.Crypt == "" {
			continue
		}

		if cryptDevs[lv.Crypt] {
			fatalf("duplicate crypt device name: %s", lv.Crypt)
		}
		cryptDevs[lv.Crypt] = true

	retryOpen:
		if len(password) == 0 {
			password = askSecret("crypt password")

			if len(password) == 0 {
				fatalf("empty password given")
			}
		}

		dev := "/dev/" + vg.VG + "/" + lv.Name

		needFormat := !devInitialized(dev)
		if needFormat {
			if !passwordVerified {
			retry:
				p2 := askSecret("verify crypt password")

				eq := bytes.Equal(password, p2)

				for i := range p2 {
					p2[i] = 0
				}

				if !eq {
					log.Print("passwords don't match")
					goto retry
				}
			}

			log.Print("formatting encrypted device ", dev)
			cmd := exec.Command("cryptsetup", "luksFormat", dev, "--key-file=-")
			cmd.Stdin = bytes.NewBuffer(password)
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			err := cmd.Run()
			if err != nil {
				fatalf("failed luksFormat: %v", err)
			}
		}

		log.Print("openning encrypted device ", lv.Crypt, " from ", dev)
		cmd := exec.Command("cryptsetup", "open", dev, lv.Crypt, "--key-file=-")
		cmd.Stdin = bytes.NewBuffer(password)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		err := cmd.Run()
		if err != nil {
			// maybe the password is wrong
			for i := range password {
				password[i] = 0
			}
			password = password[:0]
			passwordVerified = false
			goto retryOpen
		}

		if needFormat {
			zeroDevStart("/dev/mapper/" + lv.Crypt)
		}

		passwordVerified = true
	}

	for i := range password {
		password[i] = 0
	}
}

func devInitialized(dev string) bool {
	f, err := os.Open(dev)
	if err != nil {
		fatalf("failed to open %s: %v", dev, err)
	}

	defer f.Close()

	ba := make([]byte, 8192)
	_, err = f.Read(ba)
	if err != nil {
		fatalf("failed to read %s: %v", dev, err)
	}

	for _, b := range ba {
		if b != 0 {
			return true
		}
	}
	return false
}

func setupFS(vg config.LvmVG) {
	for _, lv := range vg.LVs {
		dev := "/dev/" + vg.VG + "/" + lv.Name

		if lv.Crypt != "" {
			dev = "/dev/mapper/" + lv.Crypt
		}

		if devInitialized(dev) {
			log.Print("device ", dev, " already formatted")
			continue
		}

		if lv.FS == "" {
			lv.FS = vg.Defaults.FS
		}

		log.Print("formatting ", dev, " (", lv.FS, ")")
		args := make([]string, 0)

		switch lv.FS {
		case "btrfs":
			args = append(args, "-f")
		case "ext4":
			args = append(args, "-F")
		}

		run("mkfs."+lv.FS, append(args, dev)...)
	}
}
