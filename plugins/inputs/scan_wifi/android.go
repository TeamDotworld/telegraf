package scanwifi

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AndroidCell struct {
	MAC           string `json:"bssid,omitempty"`
	Encryption    string `json:"capabilities,omitempty"`
	Channel       string `json:"channel_width,omitempty"`
	Frequency     string `json:"frequency,omitempty"`
	EncryptionKey bool   `json:"is_passpoint_network,omitempty"`
	SignalLevel   string `json:"level,omitempty"`
	ESSID         string `json:"ssid,omitempty"`
	NetworkTime   int64  `json:"timestamp,omitempty"`
	VenueName     string `json:"venue_name,omitempty"`
	Quality       int    `json:"signal_quality,omitempty"`
}

func GetAdnroidWifi() ([]byte, string) {
	var IpAddress string
	getInterface := Interface()
	if getInterface == nil {
		return nil, IpAddress
	}
	var cells []AndroidCell
	for _, interfaceName := range getInterface {
		cellsTemp, err := AndroidWifiScan(interfaceName)
		if err != nil {
			return []byte{}, ""
		}
		if len(cellsTemp) > 0 {
			cells = cellsTemp
			IpAddress = GetIPAddress(interfaceName)
			break
		}
	}

	jsonmarshal, err := json.Marshal(cells)
	if err != nil {
		return nil, IpAddress
	}
	return jsonmarshal, IpAddress
}
func Interface() []string {
	var Interface []string
	files, err := ioutil.ReadDir("/sys/class/ieee80211/")
	if err != nil {
		return nil
	}
	if len(files) == 0 {
		return nil
	}
	for _, f := range files {
		filephy, err := ioutil.ReadDir("/sys/class/ieee80211/" + f.Name())
		if err != nil {
			return nil
		}
		for _, fphy := range filephy {
			if fphy.Name() == "device" {
				file, err := ioutil.ReadDir("/sys/class/ieee80211/" + f.Name() + "/" + fphy.Name())
				if err != nil {
					return nil
				}
				for _, fi := range file {
					if fi.Name() == "net" {
						readnet, err := ioutil.ReadDir("/sys/class/ieee80211/" + f.Name() + "/" + fphy.Name() + "/" + fi.Name() + "/")
						if err != nil {
							return nil
						}
						for _, r := range readnet {
							Interface = append(Interface, r.Name())
						}
					}
				}
			}
		}
	}
	return Interface
}
func GetIPAddress(interfaceName string) string {
	var IPAddress string
	cmd, err := exec.Command("ip", "addr", "show", interfaceName).Output()
	if err != nil {
		return ""
	} else {
		scanner := bufio.NewScanner(strings.NewReader(string(cmd)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "inet") {
				splitip := strings.Split(line, " ")
				for _, s := range splitip {
					if strings.Contains(s, "192.168") {
						IPAddress = strings.Split(s, "/")[0]
						IPAddress = strings.TrimSpace(IPAddress)
						break
					}
				}
			}
		}
	}
	return IPAddress
}

func AndroidWifiScan(interfaceName string) ([]AndroidCell, error) {
	cmd := exec.Command("iw", "dev", interfaceName, "scan")
	out, _ := cmd.CombinedOutput()
	return andparse(string(out), interfaceName)
}

func andparse(input string, interfaceName string) (cells []AndroidCell, err error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	cells = []AndroidCell{}
	wificell := AndroidCell{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "BSS") && strings.Contains(line, interfaceName) {
			// re for mac
			re := regexp.MustCompile(`([0-9a-fA-F]{2}[:-]){5}([0-9a-fA-F]{2})`)
			mac := re.FindString(line)
			wificell.MAC = mac
			wificell.NetworkTime = time.Now().Unix()
		} else if strings.Contains(line, "SSID:") {
			if len(strings.Fields(line)) > 1 {
				if strings.Contains(strings.Fields(line)[1], "\\x00") {
					wificell.ESSID = "Unidentified"
				} else if strings.Contains(strings.Fields(line)[1], "\\x") {
					emoji, _ := hex.DecodeString(strings.Replace(strings.Fields(line)[1], "\\x", "", -1))
					wificell.ESSID = strings.ReplaceAll(string(emoji), "\n", "")
				} else {
					wificell.ESSID = strings.Fields(line)[1]
				}
			} else {
				wificell.ESSID = "Unidentified"
			}
		} else if strings.Contains(line, "signal:") {
			wificell.SignalLevel = strings.Fields(line)[1]
		} else if strings.Contains(line, "freq:") {
			freq := strings.Fields(line)[1]
			freqint, _ := strconv.Atoi(freq)
			if err == nil {
				wificell.Frequency = fmt.Sprintf("%.2f", float64(freqint)/1000)
			}

		} else if strings.Contains(line, "primary channel:") {
			wificell.Channel = strings.Fields(line)[1]
		}
		if wificell.MAC != "" && wificell.ESSID != "" && wificell.SignalLevel != "" && wificell.Frequency != "" && wificell.Channel != "" {
			cells = append(cells, wificell)
			wificell = AndroidCell{}
		}
	}
	return cells, scanner.Err()
}
