//go:build windows
// +build windows

package device_master

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf/plugins/inputs/device_master/wmi_win"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type BIOSe struct {
	BaseBoardManufacturer string `json:"baseboard_manufacturer"`
	BaseBoardProduct      string `json:"baseboard_product"`
	BaseBoardVersion      string `json:"baseboard_version"`
	BiosReleaseDate       string `json:"bios_release_date"`
	BiosVersion           string `json:"bios_version"`
	BiosVendor            string `json:"bios_vendor"`
	SystemFamily          string `json:"system_family"`
	SystemManufacturer    string `json:"system_manufacturer"`
	SystemProductName     string `json:"system_product_name"`
	SystemVersion         string `json:"system_version"`
	SystemSKU             string `json:"system_sku"`
}

type DEVMODE struct {
	DmDeviceName       [CCHDEVICENAME]uint16
	DmSpecVersion      uint16
	DmDriverVersion    uint16
	DmSize             uint16
	DmDriverExtra      uint16
	DmFields           uint32
	DmOrientation      int16
	DmPaperSize        int16
	DmPaperLength      int16
	DmPaperWidth       int16
	DmScale            int16
	DmCopies           int16
	DmDefaultSource    int16
	DmPrintQuality     int16
	DmColor            int16
	DmDuplex           int16
	DmYResolution      int16
	DmTTOption         int16
	DmCollate          int16
	DmFormName         [CCHFORMNAME]uint16
	DmLogPixels        uint16
	DmBitsPerPel       uint32
	DmPelsWidth        uint32
	DmPelsHeight       uint32
	DmDisplayFlags     uint32
	DmDisplayFrequency uint32
	DmICMMethod        uint32
	DmICMIntent        uint32
	DmMediaType        uint32
	DmDitherType       uint32
	DmReserved1        uint32
	DmReserved2        uint32
	DmPanningWidth     uint32
	DmPanningHeight    uint32
}

const (
	CCHDEVICENAME                 = 32
	CCHFORMNAME                   = 32
	ENUM_CURRENT_SETTINGS  uint32 = 0xFFFFFFFF
	ENUM_REGISTRY_SETTINGS uint32 = 0xFFFFFFFE
	DISP_CHANGE_SUCCESSFUL uint32 = 0
	DISP_CHANGE_RESTART    uint32 = 1
	DISP_CHANGE_FAILED     uint32 = 0xFFFFFFFF
	DISP_CHANGE_BADMODE    uint32 = 0xFFFFFFFE
)

type Win32_OperatingSystem struct {
	BootDevice                                string
	BuildNumber                               string
	BuildType                                 string
	Caption                                   string
	CodeSet                                   string
	CountryCode                               string
	CreationClassName                         string
	CSCreationClassName                       string
	CSDVersion                                string
	CSName                                    string
	DataExecutionPrevention_Available         bool
	DataExecutionPrevention_32BitApplications bool
	DataExecutionPrevention_Drivers           bool
	DataExecutionPrevention_SupportPolicy     uint8
	Debug                                     bool
	Description                               string
	Distributed                               bool
	EncryptionLevel                           uint32
	FreePhysicalMemory                        uint64
	FreeSpaceInPagingFiles                    uint64
	FreeVirtualMemory                         uint64
	InstallDate                               string
	LargeSystemCache                          uint32
	LastBootUpTime                            string
	LocalDateTime                             string
	Locale                                    string
	Manufacturer                              string
	MaxNumberOfProcesses                      uint32
	MaxProcessMemorySize                      uint64
	MUILanguages                              []string
	Name                                      string
	NumberOfLicensedUsers                     uint32
	NumberOfProcesses                         uint32
	NumberOfUsers                             uint32
	OperatingSystemSKU                        uint32
	Organization                              string
	OSArchitecture                            string
	OSLanguage                                uint32
	OSProductSuite                            uint32
	OSType                                    uint16
	OtherTypeDescription                      string
	PAEEnabled                                bool
	PlusProductID                             string
	PlusVersionNumber                         string
	PortableOperatingSystem                   bool
	Primary                                   bool
	ProductType                               uint32
	RegisteredUser                            string
	SerialNumber                              string
	ServicePackMajorVersion                   uint16
	ServicePackMinorVersion                   uint16
	SizeStoredInPagingFiles                   uint64
	Status                                    string
	SuiteMask                                 uint32
	SystemDevice                              string
	SystemDirectory                           string
	SystemDrive                               string
	TotalSwapSpaceSize                        uint64
	TotalVirtualMemorySize                    uint64
	TotalVisibleMemorySize                    uint64
	Version                                   string
	WindowsDirectory                          string
	ForegroundApplicationBoost                uint8
	CurrentTimeZone                           string
}
type Win32_DesktopMonitor struct {
	Availability                uint16
	Bandwidth                   uint32
	Caption                     string
	ConfigManagerErrorCode      uint32
	ConfigManagerUserConfig     bool
	CreationClassName           string
	Description                 string
	DeviceID                    string
	DisplayType                 uint16
	ErrorCleared                bool
	ErrorDescription            string
	IsLocked                    bool
	LastErrorCode               uint32
	MonitorManufacturer         string
	MonitorType                 string
	Name                        string
	PixelsPerXLogicalInch       uint32
	PixelsPerYLogicalInch       uint32
	PNPDeviceID                 string
	PowerManagementCapabilities []uint16
	PowerManagementSupported    bool
	ScreenHeight                uint32
	ScreenWidth                 uint32
	Status                      string
	StatusInfo                  uint16
	SystemCreationClassName     string
	SystemName                  string
	InstallDate                 string
}
type DISPLAY struct {
	Resolution  string `json:"resolution"`
	Pixel       string `json:"pixel"`
	Orientation string `json:"orientation"`
	Flags       string `json:"flags"`
	frequency   string `json:"frequency"`
}

func GetWindowsMetaData() DeviceMaster {
	getbios := GETBIOSINFO()
	var Metadata DeviceMaster
	Metadata.Board.BoardName = getbios.BaseBoardProduct
	Metadata.Board.BoardVersion = getbios.BaseBoardVersion
	Metadata.Board.BoardVendor = getbios.BaseBoardManufacturer
	Metadata.Other.Manufacturer = getbios.SystemManufacturer
	Metadata.Product.ProductName = getbios.SystemProductName
	Metadata.Product.ProductFamily = getbios.SystemFamily
	Metadata.Product.ProductSKU = getbios.SystemSKU
	win32_os := GetOperatingSystem()

	if len(win32_os) > 0 {
		Metadata.Other.OsRelease = win32_os[0].Caption
		Metadata.Other.OsVersion = win32_os[0].Version
		Metadata.Other.Language = win32_os[0].MUILanguages[0]
		Metadata.Product.ProductUUID = win32_os[0].SerialNumber
		Metadata.Other.HostName = win32_os[0].CSName
		Metadata.Other.OperatingSystem = win32_os[0].OSArchitecture
	} else {
		Metadata.Other.OsRelease = "none"
		Metadata.Other.OsVersion = "none"
		Metadata.Other.Language = "en"
	}
	monitor_info := GetWin32DesktopMonitor()
	if len(monitor_info) > 0 {
		for _, monitor := range monitor_info {
			Metadata.Other.Display += monitor.Name + ","
		}
		Metadata.Other.Display = strings.TrimSuffix(Metadata.Other.Display, ",")
	}
	Metadata.Other.Kernal = runtime.GOARCH
	Metadata.Other.OtgSupport = true
	serialnumber, err := exec.Command("wmic", "bios", "get", "serialnumber").Output()
	if err != nil {
		Metadata.Product.ProductSerial = "unknown"
	} else {
		serialnumbers := strings.Replace(string(serialnumber), "SerialNumber", "", -1)
		serialnumbers = strings.Replace(strings.TrimSpace(serialnumbers), "\n", "", -1)
		Metadata.Product.ProductSerial = serialnumbers
	}
	currentUser, err := user.Current()
	if err == nil {
		if strings.Contains(currentUser.Username, "\\") {
			Metadata.Other.UserName = strings.Split(currentUser.Username, "\\")[1]
		} else {
			Metadata.Other.UserName = currentUser.Username
		}
	}
	dt := time.Now()
	Metadata.Other.DateTime = dt.Format(time.RFC3339)
	zone, offset := time.Now().Zone()
	Metadata.Other.TimeZone = zone + " " + strconv.Itoa(offset)
	Metadata.Other.Rooted = amAdmin()
	m := make(map[string]string)
	mont := GetMonitorInfo()
	err = json.Unmarshal([]byte(mont), &m)
	if err == nil {
		Metadata.Other.ScreenResolution = m["resolution"]
		Metadata.Other.Orientation = m["orientation"]
	}
	Metadata.Other.TotolProcessor = GetTotalProcessor()
	return Metadata
}
func GetTotalProcessor() int {
	process := GetCPUProcessor()
	reg_for_int := regexp.MustCompile(`[0-9]+`)
	processors := reg_for_int.FindString(process)
	processor, err := strconv.Atoi(processors)
	if err != nil {
		return 0
	}
	return processor
}

func GetCPUProcessor() string {
	key, _ := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, registry.QUERY_VALUE)
	defer key.Close()
	value, _, _ := key.GetStringValue("NUMBER_OF_PROCESSORS")
	return value
}

func amAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

func GetMonitorInfo() []byte {
	var Display DISPLAY
	user32dll := syscall.NewLazyDLL("user32.dll")
	procEnumDisplaySettingsW := user32dll.NewProc("EnumDisplaySettingsW")
	// procChangeDisplaySettingsW := user32dll.NewProc("ChangeDisplaySettingsW")

	// Get the display information.
	devMode := new(DEVMODE)
	ret, _, _ := procEnumDisplaySettingsW.Call(uintptr(unsafe.Pointer(nil)),
		uintptr(ENUM_CURRENT_SETTINGS), uintptr(unsafe.Pointer(devMode)))
	if ret == 0 {
		fmt.Println("Couldn't extract display settings.")
	} else {
		Display.Resolution = fmt.Sprintf("%d x %d", devMode.DmPelsWidth, devMode.DmPelsHeight)
		Display.Pixel = fmt.Sprintf("%d", devMode.DmBitsPerPel)
		Display.Orientation = fmt.Sprintf("%d", devMode.DmOrientation)
		Display.Flags = fmt.Sprintf("%x", devMode.DmDisplayFlags)
		Display.frequency = fmt.Sprintf("%d", devMode.DmDisplayFrequency)
		// fmt.Println(devMode.DmDeviceName, devMode.DmFormName)
	}

	jsonmarsh, err := json.Marshal(Display)
	if err != nil {
		return []byte{}
	}
	return jsonmarsh
}

func GETBIOSINFO() BIOSe {
	var bios BIOSe
	key, _ := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, registry.QUERY_VALUE)
	defer key.Close()
	// value, _, _ := key.GetStringValue(Key)
	bios.BaseBoardManufacturer, _, _ = key.GetStringValue("BaseBoardManufacturer")
	bios.BaseBoardProduct, _, _ = key.GetStringValue("BaseBoardProduct")
	bios.BaseBoardVersion, _, _ = key.GetStringValue("BaseBoardVersion")
	bios.BiosReleaseDate, _, _ = key.GetStringValue("BIOSReleaseDate")
	bios.BiosVersion, _, _ = key.GetStringValue("BIOSVersion")
	bios.BiosVendor, _, _ = key.GetStringValue("BIOSVendor")
	bios.SystemFamily, _, _ = key.GetStringValue("SystemFamily")
	bios.SystemManufacturer, _, _ = key.GetStringValue("SystemManufacturer")
	bios.SystemProductName, _, _ = key.GetStringValue("SystemProductName")
	bios.SystemVersion, _, _ = key.GetStringValue("SystemVersion")
	bios.SystemSKU, _, _ = key.GetStringValue("SystemSKU")
	return bios
}

func GetOperatingSystem() []Win32_OperatingSystem {
	var dstOperatingSystem []Win32_OperatingSystem
	q := wmi_win.CreateQuery(&dstOperatingSystem, "")
	wmi_win.Query(q, &dstOperatingSystem)
	return dstOperatingSystem
}

func GetWin32DesktopMonitor() []Win32_DesktopMonitor {
	var dstDesktopMonitor []Win32_DesktopMonitor
	q := wmi_win.CreateQuery(&dstDesktopMonitor, "")
	wmi_win.Query(q, &dstDesktopMonitor)
	return dstDesktopMonitor
}

func LinuxSystemMetrics() DeviceMaster {
	return DeviceMaster{}
}
