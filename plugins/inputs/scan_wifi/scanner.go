package scanwifi

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  interface = "wlan0"
`

type AvailableWifi struct {
	Interface     string `toml:"interface"`
	MAC           string
	Encryption    string
	Channel       int
	Frequency     float32
	EncryptionKey bool
	SignalLevel   int
	ESSID         string
	NetworkTime   int64
	VenueName     string
	Quality       int
}

func (wifi *AvailableWifi) SampleConfig() string {
	return sampleConfig
}

func (wifi *AvailableWifi) Description() string {
	return "wifi scan go-plugin for Telegraf"
}
func (wifi *AvailableWifi) Gather(acc telegraf.Accumulator) error {
	platform := GETPLATFORM()
	switch platform {
	case "linux":
		available_wifilist, err := Scan(wifi.Interface)
		if err != nil {
			acc.AddError(err)
		} else {
			for _, list := range available_wifilist {
				acc.AddFields("scan_wifi", map[string]interface{}{
					"bssid":                list.MAC,
					"capabilities":         list.Encryption,
					"channel_width":        list.Channel,
					"frequency":            list.Frequency,
					"is_passpoint_network": list.EncryptionKey,
					"level":                list.SignalLevel,
					"timestamp":            list.NetworkTime,
					"venue_name":           list.VenueName,
					"signal_quality":       list.Quality,
				}, map[string]string{
					"ssid": list.ESSID,
				})
			}
		}
	case "android":
		cellsTemp, err := AndroidWifiScan(wifi.Interface)
		if err != nil {
			acc.AddError(err)
		} else {
			for _, list := range cellsTemp {
				acc.AddFields("scan_wifi", map[string]interface{}{
					"bssid":                list.MAC,
					"capabilities":         list.Encryption,
					"channel_width":        list.Channel,
					"frequency":            list.Frequency,
					"is_passpoint_network": list.EncryptionKey,
					"level":                list.SignalLevel,
					"timestamp":            list.NetworkTime,
					"venue_name":           list.VenueName,
					"signal_quality":       list.Quality,
				}, map[string]string{
					"ssid": list.ESSID,
				})
			}
		}
	case "windows":
		wifilist, err := WinScan()
		if err != nil {
			acc.AddError(err)
		} else {
			for _, list := range wifilist {
				acc.AddFields("scan_wifi", map[string]interface{}{
					"bssid":         list.MAC,
					"capabilities":  list.Encryption,
					"channel_width": list.Channel,
					"frequency":     list.Frequency,
					"level":         list.RSSI,
					"timestamp":     list.NetworkTime,
				}, map[string]string{
					"ssid": list.SSID,
				})
			}
		}
	}
	return nil
}
func init() {
	inputs.Add("scan_wifi", func() telegraf.Input {
		return &AvailableWifi{}
	})
}

func GETPLATFORM() string {
	var OS_TYPE string
	if runtime.GOOS == "linux" {
		if !VerifyAppInstalled("getprop") {
			OS_TYPE = "linux"
		} else {
			execProp, err := exec.Command("getprop", "ro.product.board").Output()
			if err != nil {
				return ""
			}
			Platform := strings.TrimSuffix(string(execProp), "\n")
			if Platform != "" {
				OS_TYPE = "android"
			} else {
				OS_TYPE = "linux"
			}
		}
	} else if runtime.GOOS == "windows" {
		OS_TYPE = "windows"
	}
	return OS_TYPE
}

func VerifyAppInstalled(pkg string) bool {
	cmd, err := exec.Command("which", pkg).Output()
	if err != nil {
		return false
	}
	var output bool
	if len(cmd) > 0 && !strings.Contains(string(cmd), "not found") {
		output = true
	} else {
		output = false
	}
	return output
}
