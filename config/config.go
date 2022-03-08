package config

type Config struct {
	AntiPhishingCode string `json:"anti_phishing_code"`

	Keymap  string
	Modules string

	Auths []Auth

	Networks []struct {
		Name       string
		Interfaces []struct {
			Var     string
			N       int
			Regexps []string
		}
		Script string
	}

	LVM       []LvmVG
	Bootstrap Bootstrap
}

type Auth struct {
	Name     string
	SSHKey   string `yaml:"sshKey"`
	Password string `yaml:"password"`
}

type LvmVG struct {
	VG  string
	PVs struct {
		N       int
		Regexps []string
	}

	Defaults struct {
		FS   string
		Raid *RaidConfig
	}

	LVs []struct {
		Name    string
		Crypt   string
		FS      string
		Raid    *RaidConfig
		Size    string
		Extents string
	}
}

type RaidConfig struct {
	Mirrors int
	Stripes int
}

type Bootstrap struct {
	Dev  string
	Seed string
}
