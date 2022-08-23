package androidapps

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type Androidapp struct {
	PackageName string
	Version     int
	AppName     string
	ClassName   string
	VersionName string
	DataDir     string
	UserInstall bool
}

func (apps *Androidapp) SampleConfig() string {
	return sampleConfig
}
func (apps *Androidapp) Description() string {
	return "Android apps go-plugin for Telegraf"
}

func (apps *Androidapp) Gather(acc telegraf.Accumulator) error {

	host, err := os.Hostname()
	if err != nil {
		return err
	}
	if GETPLATFORM() == "android" {
		app_list := GetUserInstalledApplication(acc)
		for _, app := range app_list {
			var err error
			apps.PackageName = app
			apps.UserInstall = true
			appinfo := GetPackageInfo(app)
			apps.VersionName = appinfo.VersionName
			apps.Version, err = strconv.Atoi(appinfo.VersionCode)
			if err != nil {
				acc.AddError(err)
			}
			apps.DataDir = appinfo.DataDir
			splitappname := strings.Split(app, ".")
			for i := 1; i < len(splitappname); i++ {
				apps.AppName = apps.AppName + splitappname[i] + "."
			}
			apps.AppName = strings.TrimSuffix(apps.AppName, ".")
			acc.AddFields("android_apps", map[string]interface{}{
				"package_name":      apps.PackageName,
				"version":           apps.Version,
				"name":              apps.AppName,
				"class_name":        apps.ClassName,
				"version_name":      apps.VersionName,
				"data_dir":          apps.DataDir,
				"is_user_installed": apps.UserInstall,
			}, map[string]string{
				"source": host,
				"data":   "apps",
			})
			apps.AppName = ""
		}
	} else {
		acc.AddError(fmt.Errorf("this plugin only supported on android device."))
	}
	return nil
}

func init() {
	inputs.Add("android_apps", func() telegraf.Input {
		return &Androidapp{}
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

func GetUserInstalledApplication(acc telegraf.Accumulator) []string {
	var Apklist []string
	getapklist, err := exec.Command("pm", "list", "packages", "-3").Output()
	if err != nil {
		acc.AddError(err)
	}
	splitline := strings.Split(string(getapklist), "\n")
	for _, line := range splitline {
		if strings.Contains(line, "package:") {
			Apklist = append(Apklist, strings.Split(line, "package:")[1])
		}
	}
	return Apklist
}

type packageInfo struct {
	VersionCode string
	VersionName string
	DataDir     string
}

func GetPackageInfo(packageName string) packageInfo {
	getpackageinfo, err := exec.Command("dumpsys", "package", packageName).Output()
	if err != nil {
		return packageInfo{}
	}
	// get app version
	var pkg packageInfo

	splitline := strings.Split(string(getpackageinfo), "\n")
	for _, line := range splitline {

		switch {
		case strings.Contains(line, "versionName"):
			pkg.VersionName = strings.Split(line, "versionName=")[1]
		case strings.Contains(line, "versionCode"):
			splitspace := strings.Split(line, " ")
			for _, space := range splitspace {
				if strings.Contains(space, "versionCode") {
					pkg.VersionCode = strings.Split(space, "versionCode=")[1]
					break
				}
			}
		case strings.Contains(line, "dataDir"):
			pkg.DataDir = strings.Split(line, "dataDir=")[1]
		}
	}
	return pkg
}
