package lvm

type LVSReport struct {
	Report []struct {
		LV []LV `json:"lv"`
	} `json:"report"`
}

type LV struct {
	Name   string `json:"lv_name"`
	VGName string `json:"vg_name"`
}

func (r LVSReport) LVs() (ret []LV) {
	for _, rep := range r.Report {
		ret = append(ret, rep.LV...)
	}
	return
}
