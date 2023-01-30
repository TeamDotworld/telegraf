package cpu_temp

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type Temperature struct {
	Temperature float64
}

func (e *Temperature) SampleConfig() string {
	return sampleConfig
}

func (e *Temperature) Description() string {
	return "CPU telegraf go-plugin"
}

func (temp *Temperature) Gather(acc telegraf.Accumulator) error {
	switch GETPLATFORM() {
	case "linux":
		temp.Temperature = GetTemperature("cpu")
	case "android":
		tempCmd := exec.Command("cat", "/sys/class/thermal/thermal_zone0/temp")
		tempOut, err := tempCmd.Output()
		if err != nil {
			fmt.Println("failed to get cpu temperature", err)
		}
		temps := strings.TrimSuffix(string(tempOut), "\n")
		tempa, err := strconv.ParseFloat(temps, 64)
		if err != nil {
			fmt.Println("failed to parse cpu temperature", err)
		}
		temp.Temperature = tempa / 1000
	case "windows":
		regex := `([0-9]+)`
		re := regexp.MustCompile(regex)
		gettemperature, err := exec.Command("powershell", "wmic", "cpu", "get", "loadpercentage").Output()
		if err != nil {
			fmt.Println("Error getting cpu temperature", err)
		} else {
			gettemperatures := strings.TrimSuffix(string(gettemperature), "\r\n")
			gettemperatures = strings.TrimSpace(gettemperatures)
			temperaturematch := re.FindStringSubmatch(gettemperatures)
			if len(temperaturematch) > 0 {
				temp.Temperature, _ = strconv.ParseFloat(strings.TrimSpace(temperaturematch[0]), 64)
			}
		}
	case "darwin":
		f, err := strconv.ParseFloat(GetCPUtemp(), 64)
		if err != nil {
			acc.AddError(err)
		}
		temp.Temperature = f
	}
	if temp.Temperature > 0 {
		acc.AddFields("cpu_temp", map[string]interface{}{
			"cpu_temp": temp.Temperature,
		}, map[string]string{
			"cpu": "temp",
		})
	}
	return nil
}

func init() {
	inputs.Add("cpu_temp", func() telegraf.Input {
		return &Temperature{}
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
	} else {
		OS_TYPE = runtime.GOOS
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

func GetTemperature(needTemperature string) float64 {
	Thermal_Directory := "/sys/class/thermal/"
	Thermal_Zone := "thermal_zone"
	Thermal_Type := "type"
	Thermal_Temp := "temp"
	if _, err := os.Stat(Thermal_Directory); err != nil {
		fmt.Println("failed to stat thermal directory", err)
		return 0
	}
	readDirectory, err := ioutil.ReadDir(Thermal_Directory)
	if err != nil {
		fmt.Println("Error in reading directory")
	}
	if len(readDirectory) > 0 {
		for _, file := range readDirectory {
			if strings.Contains(file.Name(), Thermal_Zone) {
				fileName := Thermal_Directory + file.Name() + "/" + Thermal_Type
				fileType, err := ioutil.ReadFile(fileName)
				if err != nil {
					fmt.Println("Error in reading file")
				}
				fileName = Thermal_Directory + file.Name() + "/" + Thermal_Temp
				fileTemp, err := ioutil.ReadFile(fileName)
				if err != nil {
					fmt.Println("Error in reading file")
				}
				tempInt, err := strconv.Atoi(strings.TrimSuffix(string(fileTemp), "\n"))
				if err != nil {
					fmt.Println("Error in converting to int", err)
				}
				tempFloat := float64(tempInt) / 1000
				if needTemperature == "cpu" {
					if strings.Contains(string(fileType), "temp") {
						return tempFloat
					}
				} else if needTemperature == "acpi" {
					if strings.Contains(string(fileType), "acpi") {
						return tempFloat
					}
				}
				return tempFloat
			}
		}
	}
	return 0
}
