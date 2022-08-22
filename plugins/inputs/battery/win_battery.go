//go:build windows
// +build windows

package battery

import (
	"fmt"
	"github.com/influxdata/telegraf/plugins/inputs/battery/wmi"
	"math"
	"os/exec"
	"regexp"
	"strconv"
)

var (
	batterylevelmatch uint16
	statusmatch       uint16
	voltagematch      uint64
	chemistrymatch    uint16
	temperaturematch  []string
)

func GetBattery(platform string) Battery {
	var batterystruct Battery
	batteryinfo := GetWin32Battery()
	regex := `([0-9]+)`
	re := regexp.MustCompile(regex)
	if len(batteryinfo) > 0 {
		batterylevelmatch = batteryinfo[0].EstimatedChargeRemaining
		statusmatch = batteryinfo[0].BatteryStatus
		voltagematch = batteryinfo[0].DesignVoltage
		chemistrymatch = batteryinfo[0].Chemistry
	}
	gettemperature, err := exec.Command("powershell", "\"Get-CimInstance -ClassName Win32_TemperatureProbe | Select-Object -Property DeviceID, Accuracy\"").Output()
	if err == nil {
		temperaturematch = re.FindStringSubmatch(string(gettemperature))
	}

	// batterystruct
	batterystruct.Level = int(batterylevelmatch)
	if batterystruct.Level > 100 {
		batterystruct.Health = "Full"
	} else if batterystruct.Level > 80 {
		batterystruct.Health = "Good"
	} else if batterystruct.Level > 60 {
		batterystruct.Health = "OK"
	} else if batterystruct.Level > 40 {
		batterystruct.Health = "Low"
	} else if batterystruct.Level > 20 {
		batterystruct.Health = "Very Low"
	} else {
		batterystruct.Health = "Critical"
	}

	switch statusmatch {
	case 1:
		batterystruct.Status = "Battery Power"
		batterystruct.Plugged = "None"
	case 2:
		batterystruct.Status = "AC Power"
		batterystruct.Plugged = "AC"
	case 3:
		batterystruct.Status = "Fully Charged"
		batterystruct.Plugged = "AC"
	case 4:
		batterystruct.Status = "Low"
		batterystruct.Plugged = "None"
	case 5:
		batterystruct.Status = "Critical"
		batterystruct.Plugged = "None"
	case 6:
		batterystruct.Status = "Charging"
		batterystruct.Plugged = "AC"
	case 7:
		batterystruct.Status = "Charging and High"
		batterystruct.Plugged = "AC"
	case 8:
		batterystruct.Status = "Charging and Low"
		batterystruct.Plugged = "AC"
	case 9:
		batterystruct.Status = "Charging and Critical"
		batterystruct.Plugged = "AC"
	case 10:
		batterystruct.Status = "Undefined"
		batterystruct.Plugged = "None"
	case 11:
		batterystruct.Status = "Partially Charged"
		batterystruct.Plugged = "AC"
	default:
		batterystruct.Status = "Unknown"
		batterystruct.Plugged = "None"
	}
	// convert to celsius
	voltage := voltagematch
	fvoltage := float64(voltage) / 1000
	batterystruct.Voltage = strconv.FormatFloat(fvoltage, 'f', -1, 64)
	if len(temperaturematch) > 0 {
		temp := temperaturematch[0]
		// convert to celsius
		temperature, err := strconv.ParseFloat(temp, 64)
		if err == nil {
			ftemperature := math.Round(temperature)
			if batterystruct.Voltage != "0" && batterystruct.Voltage != "" {
				batterystruct.Temperature = fmt.Sprintf("%.2f℃", ftemperature)
			}
		}
	}
	switch chemistrymatch {
	case 1:
		batterystruct.Technology = "Other"
	case 2:
		batterystruct.Technology = "Unknown"
	case 3:
		batterystruct.Technology = "LeadAcid"
	case 4:
		batterystruct.Technology = "NickelCadmium"
	case 5:
		batterystruct.Technology = "NickelMetalHydride"
	case 6:
		batterystruct.Technology = "LithiumIon"
	case 7:
		batterystruct.Technology = "Zincair"
	case 8:
		batterystruct.Technology = "LithiumPolymer"
	default:
		batterystruct.Technology = "Unknown"
	}
	return batterystruct
}

type Win32_Battery struct {
	BatteryRechargeTime      uint32
	ConfigManagerErrorCode   uint32
	DesignCapacity           uint32
	EstimatedRunTime         uint32
	ExpectedBatteryLife      uint32
	ExpectedLife             uint32
	FullChargeCapacity       uint32
	LastErrorCode            uint32
	MaxRechargeTime          uint32
	TimeOnBattery            uint32
	TimeToFullCharge         uint32
	Chemistry                uint16
	DesignVoltage            uint64
	EstimatedChargeRemaining uint16
	StatusInfo               uint16
	Availability             uint16
	BatteryStatus            uint16
	PowerManagementSupported bool
	ErrorCleared             bool
	ConfigManagerUserConfig  bool
	Caption                  string
	CreationClassName        string
	Description              string
	DeviceID                 string
	ErrorDescription         string
	Name                     string
	PNPDeviceID              string
	SmartBatteryVersion      string
	Status                   string
	SystemCreationClassName  string
	SystemName               string
	InstallDate              string
	// PowerManagementCapabilities []uint16
}

func GetWin32Battery() []Win32_Battery {
	var dstBattery []Win32_Battery
	q := wmi_win.CreateQuery(&dstBattery, "")
	_ = wmi_win.Query(q, &dstBattery)
	return dstBattery
}
