//go:build linux
// +build linux

package device_master

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	LinuxDMIPath   = "/sys/class/dmi/id/"
	LinuxProcsPath = "/proc/device-tree/"
)

func SystemMetrics() DeviceMaster {
	var all_metrics DeviceMaster
	if _, err := os.Stat(LinuxDMIPath); !os.IsNotExist(err) {
		DMIRead, err := ioutil.ReadDir(LinuxDMIPath)
		if err != nil {
			fmt.Println("Error of read: ", err)
		} else {
			for _, file := range DMIRead {
				ReadTxt := ReadTxtFile(LinuxDMIPath + file.Name())
				switch file.Name() {
				case "bios_date":
					all_metrics.Bios.BiosDate = ReadTxt
				case "bios_release":
					all_metrics.Bios.BiosRelease = ReadTxt
				case "bios_vendor":
					all_metrics.Bios.BiosVendor = ReadTxt
				case "bios_version":
					all_metrics.Bios.BiosVersion = ReadTxt
				case "board_asset_tag":
					all_metrics.Board.BoardAssetTag = ReadTxt
				case "board_name":
					all_metrics.Board.BoardName = ReadTxt
				case "board_serial":
					all_metrics.Board.BoardSerial = ReadTxt
				case "board_vendor":
					all_metrics.Board.BoardVendor = ReadTxt
				case "board_version":
					all_metrics.Board.BoardVersion = ReadTxt
				case "chassis_asset_tag":
					all_metrics.Chassis.ChassisAssetTag = ReadTxt
				case "chassis_serial":
					all_metrics.Chassis.ChassisSerial = ReadTxt
				case "chassis_type":
					all_metrics.Chassis.ChassisType = ReadTxt
				case "chassis_vendor":
					all_metrics.Chassis.ChassisVendor = ReadTxt
					all_metrics.Other.Manufacturer = ReadTxt
				case "chassis_version":
					all_metrics.Chassis.ChassisVersion = ReadTxt
				case "ec_firmware_release":
					all_metrics.Product.EcFirmwareRelease = ReadTxt
				case "product_family":
					all_metrics.Product.ProductFamily = ReadTxt
				case "product_name":
					all_metrics.Product.ProductName = ReadTxt
				case "product_serial":
					all_metrics.Product.ProductSerial = ReadTxt
				case "product_sku":
					all_metrics.Product.ProductSKU = ReadTxt
				case "product_uuid":
					all_metrics.Product.ProductUUID = ReadTxt
				case "product_version":
					all_metrics.Product.ProductVersion = ReadTxt
				case "sys_vendor":
					all_metrics.Product.SysVendor = ReadTxt
				}
			}
		}
	}
	all_metrics.Other.TimeZone = ReadTxtFile("/etc/timezone")
	all_metrics.Other.DateTime = time.Now().Format(time.RFC3339)
	all_metrics.Other.Language = os.Getenv("LANG")
	host, _ := exec.Command("hostname").Output()
	all_metrics.Other.HostName = strings.TrimSuffix(string(host), "\n")
	user, _ := user.Current()
	all_metrics.Other.UserName = user.Username
	oprelease, _ := Read()
	all_metrics.Other.OsVersion = oprelease["VERSION"]
	all_metrics.Other.OsRelease = oprelease["PRETTY_NAME"]
	conint, _ := exec.Command("grep", "-c", "processor", "/proc/cpuinfo").Output()
	conv := strings.TrimSuffix(string(string(conint)), "\n")
	all_metrics.Other.TotolProcessor, _ = strconv.Atoi(string(conv))
	if all_metrics.Other.UserName == "root" {
		all_metrics.Other.Rooted = true
		all_metrics.Other.IsAdmin = true
	} else {
		all_metrics.Other.Rooted = false
	}
	all_metrics.Other.OtgSupport = true
	all_metrics.Other.Kernal = runtime.GOARCH
	var isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z0-9_ .-^\w]*$`).MatchString
	var displays string
	direct, err := ioutil.ReadDir("/sys/class/drm/")
	if err == nil {
		for _, file := range direct {
			directory, _ := ioutil.ReadDir("/sys/class/drm/" + file.Name())
			for _, files := range directory {
				if files.Name() == "edid" {
					edid, _ := exec.Command("strings", "/sys/class/drm/"+file.Name()+"/"+files.Name()).Output()
					splitD := strings.Split(string(edid), "\n")
					for _, v := range splitD {
						v = strings.TrimSpace(v) // remove leading and trailing spaces
						if len(v) > 8 {
							if isStringAlphabetic(v) {
								displays += v + ","
							}
						}
					}
				}
			}
		}
	}
	all_metrics.Other.Display = strings.TrimSuffix(displays, ",")
	dir, err := ioutil.ReadDir("/sys/class/graphics/")
	if err == nil {
		for _, file := range dir {
			directory, err := ioutil.ReadDir("/sys/class/graphics/" + file.Name())
			if err == nil {
				for _, files := range directory {
					if files.Name() == "virtual_size" {
						virtual_size, err := ioutil.ReadFile("/sys/class/graphics/" + file.Name() + "/" + files.Name())
						if err == nil {
							all_metrics.Other.ScreenResolution = strings.TrimSuffix(strings.Replace(string(virtual_size), ",", " x ", -1), "\n")
						}
					}
				}
			}
		}
	}
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err == nil {
		all_metrics.Other.OperatingSystem = getOSBit()
	}
	return all_metrics
}

// func int8ToStr(arr []int8) string {
// 	b := make([]byte, 0, len(arr))
// 	for _, v := range arr {
// 		if v == 0x00 {
// 			break
// 		}
// 		b = append(b, byte(v))
// 	}
// 	return string(b)
// }

func getOSBit() string {
	var result string
	getOSbit, err := exec.Command("getconf", "LONG_BIT").Output()
	if err != nil {
		return strings.TrimSpace(string(getOSbit))
	}
	return result
}
