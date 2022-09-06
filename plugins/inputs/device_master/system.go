package device_master

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type BIOS struct {
	BiosDate    string
	BiosRelease string
	BiosVendor  string
	BiosVersion string
}

type BOARD struct {
	BoardAssetTag string
	BoardName     string
	BoardSerial   string
	BoardVendor   string
	BoardVersion  string
}

type CHASSIS struct {
	ChassisAssetTag string
	ChassisSerial   string
	ChassisType     string
	ChassisVendor   string
	ChassisVersion  string
}

type PRODUCT struct {
	ProductFamily     string
	ProductName       string
	ProductSerial     string
	ProductSKU        string
	ProductUUID       string
	ProductVersion    string
	EcFirmwareRelease string
	SysVendor         string
}
type OTHERS struct {
	DateTime         string
	Language         string
	TimeZone         string
	HostName         string
	UserName         string
	Manufacturer     string
	OsVersion        string
	OsRelease        string
	OperatingSystem  string
	TotolProcessor   int
	Rooted           bool
	Display          string
	ScreenResolution string
	SDKVersion       string
	Fingerprint      string
	RadioVersion     string
	Orientation      string
	IsAdmin          bool
	SecurityPatch    string
	GLVersion        string
	OtgSupport       bool
	Kernal           string
	Type             string
}

type DeviceMaster struct {
	Bios    BIOS
	Board   BOARD
	Chassis CHASSIS
	Product PRODUCT
	Other   OTHERS
}

func (master *DeviceMaster) SampleConfig() string {
	return sampleConfig
}

func (master *DeviceMaster) Description() string {
	return "This is master of device data collecter go-plugin for Telegraf"
}

func (master *DeviceMaster) Gather(acc telegraf.Accumulator) error {
	platform := GETPLATFORM()
	var result DeviceMaster
	switch platform {
	case "linux":
		result = LinuxSystemMetrics()
	case "android":
		result = GetAndroidTelemetryData()
	case "windows":
		result = GetWindowsMetaData()
	}
	acc.AddFields("system_metrics", map[string]interface{}{
		"bios_date":           result.Bios.BiosDate,
		"bios_release":        result.Bios.BiosRelease,
		"bios_vendor":         result.Bios.BiosVendor,
		"bios_version":        result.Bios.BiosVersion,
		"board_asset_tag":     result.Board.BoardAssetTag,
		"board_name":          result.Board.BoardName,
		"board_serial":        result.Board.BoardSerial,
		"board_vendor":        result.Board.BoardVendor,
		"board_version":       result.Board.BoardVersion,
		"chassis_asset_tag":   result.Chassis.ChassisAssetTag,
		"chassis_serial":      result.Chassis.ChassisSerial,
		"chassis_type":        result.Chassis.ChassisType,
		"chassis_vendor":      result.Chassis.ChassisVendor,
		"chassis_version":     result.Chassis.ChassisVersion,
		"ec_firmware_release": result.Product.EcFirmwareRelease,
		"product_family":      result.Product.ProductFamily,
		"product_name":        result.Product.ProductName,
		"product_serial":      result.Product.ProductSerial,
		"product_sku":         result.Product.ProductSKU,
		"product_uuid":        result.Product.ProductUUID,
		"product_version":     result.Product.ProductVersion,
		"sys_vendor":          result.Product.SysVendor,
		"date_time":           result.Other.DateTime,
		"time_zone":           result.Other.TimeZone,
		"hostname":            result.Other.HostName,
		"username":            result.Other.UserName,
		"os_version":          result.Other.OsVersion,
		"os_release":          result.Other.OsRelease,
		"type":                result.Other.OperatingSystem,
		"total_processor":     result.Other.TotolProcessor,
		"rooted":              result.Other.Rooted,
		"display":             result.Other.Display,
		"screen_resolution":   result.Other.ScreenResolution,
		"sdk_version":         result.Other.SDKVersion,
		"fingerprint":         result.Other.Fingerprint,
		"radio_version":       result.Other.RadioVersion,
		"orientation":         result.Other.Orientation,
		"is_admin":            result.Other.IsAdmin,
		"security_patch":      result.Other.SecurityPatch,
		"gl_version":          result.Other.GLVersion,
		"otg_support":         result.Other.OtgSupport,
	}, map[string]string{})
	return nil
}

func init() {
	inputs.Add("device_master", func() telegraf.Input {
		return &DeviceMaster{}
	})
}

func ReadTxtFile(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	var content string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += strings.TrimSpace(scanner.Text())
	}
	return content
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
