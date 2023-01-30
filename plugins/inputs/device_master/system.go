package device_master

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"howett.net/plist"
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

type MacMetrix struct {
	Items []struct {
		Name                 string `plist:"_name" json:"name,omitempty"`
		BootMode             string `plist:"boot_mode" json:"boot_mode,omitempty"`
		BootVolume           string `plist:"boot_volume" json:"boot_volume,omitempty"`
		KernalVersion        string `plist:"kernel_version" json:"kernel_version,omitempty"`
		LocalHostName        string `plist:"local_host_name" json:"local_host_name,omitempty"`
		OsVersion            string `plist:"os_version" json:"os_version,omitempty"`
		SecureVM             string `plist:"secure_vm" json:"secure_vm,omitempty"`
		IntegrityEnabled     string `plist:"system_integrity" json:"system_integrity,omitempty"`
		UpTime               string `plist:"uptime" json:"uptime,omitempty"`
		UserName             string `plist:"user_name" json:"user_name,omitempty"`
		ActivationLockStatus string `plist:"activation_lock_status" json:"activation_lock_status,omitempty"`
		BootRomVersion       string `plist:"boot_rom_version" json:"boot_rom_version,omitempty"`
		ChipType             string `plist:"chip_type" json:"chip_type,omitempty"`
		MachineModel         string `plist:"machine_model" json:"machine_model,omitempty"`
		MachineName          string `plist:"machine_name" json:"machine_name,omitempty"`
		ModelNumber          string `plist:"model_number" json:"model_number,omitempty"`
		NumberProcessors     string `plist:"number_processors" json:"number_processors,omitempty"`
		OsLoaderVersion      string `plist:"os_loader_version" json:"os_loader_version,omitempty"`
		PhysicalMemory       string `plist:"physical_memory" json:"physical_memory,omitempty"`
		PlatformUUID         string `plist:"platform_UUID" json:"platform_UUID,omitempty"`
		ProvisioningUDID     string `plist:"provisioning_UDID" json:"provisioning_UDID,omitempty"`
		SerialNumber         string `plist:"serial_number" json:"serial_number,omitempty"`
		SppciCores           string `plist:"sppci_cores" json:"total_number_of_cores,omitempty"`
		SppciModel           string `plist:"sppci_model" json:"chipset_model,omitempty"`
		DateTime             string `json:"date_time,omitempty"`
		SpdisplaysNdrvs      []*struct {
			Name                string `plist:"_name" json:"display_name,omitempty"`
			DisplaySerialNumber string `plist:"_spdisplays_display-serial-number" json:"display_serial_number,omitempty"`
			Resolution          string `plist:"_spdisplays_resolution" json:"display_resolution,omitempty"`
		} `plist:"spdisplays_ndrvs" json:"display,omitempty"`
	} `plist:"_items"`
}

func (master *DeviceMaster) SampleConfig() string {
	return sampleConfig
}

func (master *DeviceMaster) Description() string {
	return "This is master of device data collecter go-plugin for Telegraf"
}

func (master *DeviceMaster) Gather(acc telegraf.Accumulator) error {
	platform := GETPLATFORM()
	var (
		result DeviceMaster
	)
	switch platform {
	case "linux":
		result = SystemMetrics()
	case "android":
		result = GetAndroidTelemetryData()
	case "windows":
		result = GetWindowsMetaData()
	case "darwin":
		result = SystemMetrics()
	}

	if platform != "darwin" {
		Field := map[string]interface{}{
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
		}
		acc.AddFields("system_metrics", Field, map[string]string{})
	} else {
		args := []string{"-xml", "SPSoftwareDataType", "SPHardwareDataType", "SPDisplaysDataType"}
		out, err := exec.Command("system_profiler", args...).Output()
		if err != nil {
			acc.AddError(err)
		}
		var data []*MacMetrix
		if _, err = plist.Unmarshal(out, &data); err != nil {
			fmt.Println(err)
		}
		for _, b := range data {
			for _, c := range b.Items {
				c.DateTime = time.Now().Format(time.RFC3339)
				if c.SpdisplaysNdrvs != nil {
					for _, v := range c.SpdisplaysNdrvs {
						mars, err := json.Marshal(v)
						if err != nil {
							acc.AddError(err)
						}
						fields := make(map[string]interface{})
						json.Unmarshal(mars, &fields)
						acc.AddFields("system_metrics", fields, map[string]string{})
					}
				} else {
					mars, err := json.Marshal(c)
					if err != nil {
						acc.AddError(err)
					}
					fields := make(map[string]interface{})
					json.Unmarshal(mars, &fields)
					acc.AddFields("system_metrics", fields, map[string]string{})
				}
			}
		}

	}
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
