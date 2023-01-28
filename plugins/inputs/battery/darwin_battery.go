//go:build darwin
// +build darwin

package battery

import (
	plist "howett.net/plist"
	"os/exec"
)

type Darwinbattery struct {
	Amperage          int64 `plist:"Amperage"`
	CurrentCapacity   int   `plist:"CurrentCapacity" json:"current_capacity"`
	ExternalConnected bool  `plist:"ExternalConnected" json:"external_connected"`
	DesignCapacity    int   `plist:"DesignCapacity"`
	IsCharging        bool  `plist:"IsCharging" json:"is_charging"`
	MaxCapacity       int   `plist:"MaxCapacity" json:"max_capacity"`
	Voltage           int   `json:"voltage"`
	FullyCharged      bool  `plist:"FullyCharged"`
}

func GetBattery(platform string) Battery {
	var batt Battery
	out, err := exec.Command("ioreg", "-n", "AppleSmartBattery", "-r", "-a").Output()
	if err != nil {
		return Battery{}
	}

	if len(out) == 0 {
		// No batteries.
		return Battery{}
	}

	var data []*Darwinbattery
	if _, err = plist.Unmarshal(out, &data); err != nil {
		return Battery{}
	}
	for _, b := range data {
		volts := float64(b.Voltage) / 1000
		batt.Level = int(float64(b.CurrentCapacity) * volts)
		switch {
		case !b.ExternalConnected:
			batt.Status = "Discharging"
		case b.IsCharging:
			batt.Status = "Charging"
		case b.CurrentCapacity == 0:
			batt.Status = "Empty"
		case b.FullyCharged:
			batt.Status = "Full"
		default:
			batt.Status = "Unknown"
		}
	}

	return batt
}
