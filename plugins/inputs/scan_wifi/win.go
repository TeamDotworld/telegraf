package scanwifi

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Wifi struct {
	MAC         string `json:"bssid,omitempty"`
	SSID        string `json:"ssid,omitempty"`
	RSSI        int    `json:"level,omitempty"`
	Channel     int    `json:"channel_width,omitempty"`
	Frequency   string `json:"frequency,omitempty"`
	Encryption  string `json:"capabilities,omitempty"`
	NetworkTime int64  `json:"timestamp,omitempty"`
}

func WinScan() ([]Wifi, error) {
	command := "netsh.exe wlan show networks mode=Bssid"
	stdout, _, err := runCommand(10*time.Second, command)
	if err != nil {
		return []Wifi{}, err
	}
	wifidata, err := parseWindows(stdout)
	if err != nil {
		return []Wifi{}, err
	}
	return wifidata, nil
}

func parseWindows(output string) (wifis []Wifi, err error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	w := Wifi{}
	wifis = []Wifi{}
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "BSSID") {
			fs := strings.Fields(line)
			if len(fs) == 4 {
				w.MAC = fs[3]
			}
		}

		if strings.Contains(line, "SSID") && !strings.Contains(line, "BSSID") {
			w.SSID = strings.TrimSpace(strings.Split(line, ":")[1])
		}

		if strings.Contains(line, "%") {
			fs := strings.Fields(line)
			if len(fs) == 3 {
				w.RSSI, err = strconv.Atoi(strings.Replace(fs[2], "%", "", 1))
				if err != nil {
					return
				}
				w.RSSI = (w.RSSI / 2) - 100
			}
		}
		if strings.Contains(line, "Channel") {
			fs := strings.Fields(line)
			if len(fs) == 3 {
				w.Channel, err = strconv.Atoi(fs[2])
				if err != nil {
					return
				}
			}
		}
		w.NetworkTime = time.Now().Unix()

		if strings.Contains(line, "Band") {
			freq := strings.Split(line, ":")
			if len(freq) == 2 {
				w.Frequency = freq[1]
			}
		}

		if strings.Contains(line, "Encryption") {
			fs := strings.Fields(line)
			if len(fs) == 3 {
				w.Encryption = fs[2]
			}
		}

		if w.SSID != "" && w.RSSI != 0 && w.MAC != "" && w.Channel != 0 && w.Encryption != "" && w.NetworkTime != 0 && w.Frequency != "" {
			wifis = append(wifis, w)
			w = Wifi{}
		}
	}
	return
}
func runCommand(tDuration time.Duration, commands string) (stdout, stderr string, err error) {
	command := strings.Fields(commands)
	cmd := exec.Command(command[0])
	if len(command) > 0 {
		cmd = exec.Command(command[0], command[1:]...)
	}
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Start()
	if err != nil {
		return
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(tDuration):
		err = cmd.Process.Kill()
	case err = <-done:
		stdout = outb.String()
		stderr = errb.String()
	}
	return
}
