package battery

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type Battery struct {
	Health      string `json:"health"`
	Level       int    `json:"level"`
	Plugged     string `json:"plugged"`
	Status      string `json:"status"`
	Technology  string `json:"technology"`
	Temperature string `json:"temperature"`
	Voltage     string `json:"voltage"`
}

func (battery *Battery) SampleConfig() string {
	return sampleConfig
}

func (battery *Battery) Description() string {
	return "Battery go-plugin for Telegraf"
}

func (battery *Battery) Gather(acc telegraf.Accumulator) error {
	platform := GETPLATFORM()
	batteries := GetBattery(platform)

	acc.AddFields("battery",
		map[string]interface{}{
			"health":      batteries.Health,
			"level":       batteries.Level,
			"plugged":     batteries.Plugged,
			"status":      batteries.Status,
			"technology":  batteries.Technology,
			"temperature": batteries.Temperature,
			"voltage":     batteries.Voltage,
		}, map[string]string{})
	return nil
}

func Checkdevbat() string {
	getbatteryinfo, getbatterybytebuffer := exec.Command("upower", "--show-info", "/org/freedesktop/UPower/devices/DisplayDevice"), new(bytes.Buffer)
	getbatteryinfo.Stdout = getbatterybytebuffer
	getbatteryinfo.Run()
	ss := bufio.NewScanner(getbatterybytebuffer)
	var power_supply string
	for ss.Scan() {
		if strings.Contains(ss.Text(), "power supply") {
			supply_state := strings.Split(ss.Text(), " ")
			if supply_state[12] == "yes" {
				power_supply = "found"
			} else {
				power_supply = "not found"
			}
		}
	}
	return power_supply
}

func SystemMeta(path string) string {
	GetData, err := exec.Command("cat", path).Output()
	if err != nil {
		return ""
	}
	Data := strings.TrimSuffix(string(GetData), "\n")
	return Data
}

func GetTemperature(needTemperature string) string {
	var result string
	Thermal_Directory := "/sys/class/thermal/"
	Thermal_Zone := "thermal_zone"
	Thermal_Type := "type"
	Thermal_Temp := "temp"
	if _, err := os.Stat(Thermal_Directory); err != nil {
		return ""
	}
	readDirectory, err := ioutil.ReadDir(Thermal_Directory)
	if err != nil {
		return ""
	}
	if len(readDirectory) > 0 {
		for _, file := range readDirectory {
			if strings.Contains(file.Name(), Thermal_Zone) {
				fileName := Thermal_Directory + file.Name() + "/" + Thermal_Type
				fileType, err := ioutil.ReadFile(fileName)
				if err != nil {
					return ""
				}
				fileName = Thermal_Directory + file.Name() + "/" + Thermal_Temp
				fileTemp, err := ioutil.ReadFile(fileName)
				if err != nil {
					return ""
				}
				tempInt, err := strconv.Atoi(strings.TrimSuffix(string(fileTemp), "\n"))
				if err != nil {
					return ""
				}
				tempFloat := float64(tempInt) / 1000
				if needTemperature == "cpu" {
					if strings.Contains(string(fileType), "temp") {
						result = fmt.Sprintf("%.2f", tempFloat)
					}
				} else if needTemperature == "acpi" {
					if strings.Contains(string(fileType), "acpi") {
						result = fmt.Sprintf("%.2f", tempFloat)
					}
				}
				if result == "" {
					result = fmt.Sprintf("%.2f", tempFloat)
				}
			}
		}
	}
	return result
}

func init() {
	inputs.Add("battery", func() telegraf.Input {
		return &Battery{}
	})
}

var OS_TYPE string

func GETPLATFORM() string {
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
