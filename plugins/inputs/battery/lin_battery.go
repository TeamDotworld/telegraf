//go:build linux
// +build linux

package battery

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func GetBattery(platform string) Battery {
	var (
		level               string
		Battery_plugged     string
		Battery_status      string
		Battery_temperature string
		Battery_level       int
		Battery_technology  string
		Battery_voltage     string
		Health              string
		online              string
	)
	if platform == "linux" {
		if Checkdevbat() == "found" {
			files, _ := os.ReadDir("/sys/class/power_supply/")
			file_count := 0
			bat_lev := 0
			bat_full := 0
			for _, file := range files {
				if strings.HasPrefix(file.Name(), "BAT") {
					if _, err := os.Stat("/sys/class/power_supply/" + file.Name() + "/energy_now"); err == nil {
						file_count += 1
						getbl := SystemMeta("/sys/class/power_supply/" + file.Name() + "/energy_now")
						getblfull := SystemMeta("/sys/class/power_supply/" + file.Name() + "/energy_full")
						trimSbl, _ := strconv.Atoi(strings.TrimSuffix(getbl, "\n"))
						trimSblfull, _ := strconv.Atoi(strings.TrimSuffix(getblfull, "\n"))
						bat_lev += trimSbl
						bat_full += trimSblfull
					}
				}
			}
			if bat_lev == 0 {
				command, battery_bytes := exec.Command("upower", "--show-info", "/org/freedesktop/UPower/devices/DisplayDevice"), new(bytes.Buffer)
				command.Stdout = battery_bytes
				command.Run()
				battery_scanner := bufio.NewScanner(battery_bytes)
				tt := battery_scanner
				for battery_scanner.Scan() {
					if strings.Contains(tt.Text(), "percentage") {
						level = battery_scanner.Text()
					}
				}
				re := regexp.MustCompile("[0-9]+")
				lvl := re.FindAllString(level, -1)

				Battery_level, _ = strconv.Atoi(lvl[0])
			} else {
				batfullm := float64(bat_lev) / float64(bat_full)
				Battery_level, _ = strconv.Atoi(fmt.Sprintf("%.0f", batfullm*100))
			}

			if Battery_level > 90 {
				Health = "GOOD"
			} else {
				Health = "NORMAL"
			}
			Battery_technology = SystemMeta("/sys/class/power_supply/BAT0/technology")
			battery_voltage_t, _ := strconv.Atoi(SystemMeta("/sys/class/power_supply/BAT0/voltage_now"))
			battery_status_t := SystemMeta("/sys/class/power_supply/BAT0/status")
			battery_supply := SystemMeta("/sys/class/power_supply/AC/uevent")
			scanner := bufio.NewScanner(strings.NewReader(battery_supply))
			for scanner.Scan() {
				splitText := strings.Split(scanner.Text(), "=")
				if splitText[0] == "POWER_SUPPLY_NAME" {
					Battery_plugged = splitText[1]
				} else if splitText[0] == "POWER_SUPPLY_ONLINE" {
					online_status := splitText[1]
					if online_status == "1" {
						online = "Charging"
					} else if online_status == "0" {
						online = "Discharging"
					}
				}
			}

			Battery_voltage = strconv.Itoa(battery_voltage_t)[:2]
			if battery_status_t != "Unknown" {
				Battery_status = battery_status_t
			} else {
				Battery_status = online
			}
			temp_battery := GetTemperature("acpi")
			if temp_battery != "" {
				Battery_temperature = temp_battery
			}
		}
	} else if platform == "android" {
		var battery Battery
		getbattery, err := exec.Command("dumpsys", "battery").Output()
		if err != nil {
			return battery
		}
		splitline := strings.Split(string(getbattery), "\n")
		for _, line := range splitline {
			switch {
			case strings.Contains(line, "level"):
				Battery_level, _ = strconv.Atoi(strings.TrimSpace(strings.Split(line, "level:")[1]))
			case strings.Contains(line, "temperature"):
				Battery_temperature = strings.Split(line, "temperature:")[1]
			case strings.Contains(line, "voltage"):
				Battery_voltage = strings.Split(line, "voltage:")[1]
			case strings.Contains(line, "status"):
				if strings.Contains(line, "1") {
					Battery_status = "Unknown"
				} else if strings.Contains(line, "2") {
					Battery_status = "Charging"
				} else if strings.Contains(line, "3") {
					Battery_status = "Discharging"
				} else if strings.Contains(line, "4") {
					Battery_status = "Not charging"
				} else if strings.Contains(line, "5") {
					Battery_status = "Full"
				}
			case strings.Contains(line, "technology"):
				tech := strings.TrimSpace(strings.Split(line, "technology:")[1])
				if tech != "" {
					Battery_technology = tech
				} else {
					Battery_technology = "Unknown"
				}
			case strings.Contains(line, "health"):
				if strings.Contains(line, "1") {
					Health = "unknown"
				} else if strings.Contains(line, "2") {
					Health = "good"
				} else if strings.Contains(line, "3") {
					Health = "overheat"
				} else if strings.Contains(line, "4") {
					Health = "dead"
				} else if strings.Contains(line, "5") {
					Health = "over voltage"
				} else if strings.Contains(line, "6") {
					Health = "unspecified failure"
				} else if strings.Contains(line, "7") {
					Health = "cold"
				}
			case strings.Contains(line, "AC powered:"):
				if strings.Contains(line, "true") {
					Battery_plugged = "AC"
				} else {
					Battery_plugged = "Unknown"
				}
			}
		}
	}

	batteries := Battery{
		Health:      Health,
		Level:       Battery_level,
		Plugged:     Battery_plugged,
		Status:      Battery_status,
		Technology:  Battery_technology,
		Temperature: Battery_temperature,
		Voltage:     Battery_voltage,
	}
	return batteries
}
