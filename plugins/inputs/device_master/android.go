package device_master

import (
	"bufio"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

func GetAndroidTelemetryData() DeviceMaster {
	var metaData DeviceMaster
	getpropscmd, err := exec.Command("getprop").Output()
	if err != nil {
		return DeviceMaster{}
	}
	scanner := bufio.NewScanner(strings.NewReader(string(getpropscmd)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ro.build.version.release") {
			metaData.Other.OsVersion = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.build.quectelversion.release") {
			metaData.Other.OsRelease = RemoveCharofProps(line)
		} else if strings.Contains(line, "persist.sys.timezone") {
			metaData.Other.TimeZone = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.product.brand") {
			metaData.Product.ProductFamily = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.product.board") {
			metaData.Board.BoardName = RemoveCharofProps(line)
		} else if strings.Contains(line, "gsm.version.baseband") {
			metaData.Board.BoardVersion = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.build.display.id") {
			metaData.Other.Display = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.build.version.sdk") {
			metaData.Other.SDKVersion = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.serialno") {
			metaData.Product.ProductSerial = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.hardware") {
			metaData.Product.ProductName = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.build.type") {
			metaData.Other.OperatingSystem = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.build.user") {
			metaData.Other.UserName = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.build.host") {
			metaData.Other.HostName = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.build.id") {
			metaData.Product.ProductUUID = RemoveCharofProps(line)
		} else if strings.Contains(line, "ro.build.fingerprint") {
			metaData.Other.Fingerprint = RemoveCharofProps(line)
		} else if strings.Contains(line, "persist.panel.orientation") {
			orientation := RemoveCharofProps(line)
			if orientation == "0" {
				metaData.Other.Orientation = "portrait"
			} else if orientation == "1" {
				metaData.Other.Orientation = "landscape"
			} else {
				metaData.Other.Orientation = "unknown"
			}
		} else if strings.Contains(line, "ro.product.vendor.manufacturer") {
			metaData.Other.Manufacturer = RemoveCharofProps(line)
		} else if strings.Contains(line, "persist.sys.oem.otg_support") {
			if RemoveCharofProps(line) == "true" {
				metaData.Other.OtgSupport = true
			} else {
				metaData.Other.OtgSupport = false
			}
		} else if strings.Contains(line, "ro.device_owner") {
			state := RemoveCharofProps(line)
			if state == "true" {
				metaData.Other.IsAdmin = true
			} else {
				metaData.Other.IsAdmin = false
			}
		} else if strings.Contains(line, "ro.build.version.security_patch") {
			metaData.Other.SecurityPatch = RemoveCharofProps(line)
		}
	}
	metaData.Other.Kernal = runtime.GOARCH

	current_time := time.Now().Format("2006-01-02 15:04:05")
	metaData.Other.DateTime = current_time
	isRoot := VerifiPackageInstalled("su")
	metaData.Other.Rooted = isRoot
	ScreenResoltion := GetScreenResolution()
	metaData.Other.ScreenResolution = ScreenResoltion
	lang, OK := os.LookupEnv("LANG")
	if OK {
		metaData.Other.Language = lang
	} else {
		metaData.Other.Language = "en"
	}
	// Get Total Processor Count
	total_processor, err := exec.Command("cat", "/proc/cpuinfo").Output()
	if err == nil {
		splittotal_processor := strings.Split(string(total_processor), "\n")
		for _, line := range splittotal_processor {
			if strings.Contains(line, "processor") {
				metaData.Other.TotolProcessor = metaData.Other.TotolProcessor + 1
			}
		}
	}

	// Get GL Version
	gl_version, err := exec.Command("dumpsys", "SurfaceFlinger").Output()
	if err == nil {
		splittgl_version := strings.Split(string(gl_version), "\n")
		if len(splittgl_version) < 0 {
			for _, line := range splittgl_version {
				if strings.Contains(line, "GLES:") {
					// get 3.2 code
					splittgl_ver := strings.Split(line, ",")
					for _, line := range splittgl_ver {
						if strings.Contains(line, "OpenGL") {
							// get float value
							re := regexp.MustCompile(`\d+\.\d+`)
							gl_version_float := re.FindString(line)
							metaData.Other.GLVersion = gl_version_float
						}
					}
				}
			}
		}
	}
	return metaData
}

func RemoveCharofProps(props string) string {
	props = strings.Split(props, ":")[1]
	if strings.Contains(props, "[") {
		props = strings.ReplaceAll(props, "[", "")
		props = strings.ReplaceAll(props, "]", "")
		props = strings.TrimSpace(props)
	}
	return props
}

func VerifiPackageInstalled(Name string) bool {
	_, err := exec.Command("which", Name).Output()
	return err == nil
}

func GetScreenResolution() string {
	var size string
	screensizecmd, err := exec.Command("wm", "size").Output()
	if err == nil {
		splitline := strings.Split(string(screensizecmd), "\n")
		for _, line := range splitline {
			if strings.Contains(line, "Physical size:") {
				splitactivity := strings.Split(line, " ")
				for _, activity := range splitactivity {
					if strings.Contains(activity, "x") {
						size = activity
						break
					}
				}
			}
		}
	}
	return size
}
