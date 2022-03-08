package lvm

type PVSReport struct {
	Report []struct {
		PV []PV `json:"pv"`
	} `json:"report"`
}

type PV struct {
	Name   string `json:"pv_name"`
	VGName string `json:"vg_name"`
}

func (r PVSReport) PVs() (ret []PV) {
	for _, rep := range r.Report {
		ret = append(ret, rep.PV...)
	}
	return
}
