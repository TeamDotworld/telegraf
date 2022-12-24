package androidapps

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
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

type Androidapp struct {
	PackageName    string
	Version        int
	AppName        string
	ClassName      string
	VersionName    string
	DataDir        string
	UserInstall    bool
	LastUpdateTime string
	RinningTime    string
	LastTime       string
}

func (apps *Androidapp) SampleConfig() string {
	return sampleConfig
}
func (apps *Androidapp) Description() string {
	return "Android apps go-plugin for Telegraf"
}

func (apps *Androidapp) Gather(acc telegraf.Accumulator) error {
	if _, err := os.Stat("/data/hive/aapt"); err != nil {
		DownloadFile("/data/hive/aapt", "https://dothive-prod.s3.ap-southeast-1.amazonaws.com/public/aapt")
		os.Chmod("/data/hive/aapt", 0777)
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
			if appinfo.LabelName == "" {
				splitappname := strings.Split(app, ".")
				for i := 1; i < len(splitappname); i++ {
					apps.AppName = apps.AppName + splitappname[i] + "."
				}
				apps.AppName = strings.TrimSuffix(apps.AppName, ".")
			} else {
				apps.AppName = appinfo.LabelName
			}
			apps.LastUpdateTime = appinfo.LastUpdateTime
			apps.RinningTime, apps.LastTime = GetUsageStats(apps.PackageName)
			acc.AddFields("android_apps", map[string]interface{}{
				"package_name":      apps.PackageName,
				"version":           apps.Version,
				"name":              apps.AppName,
				"class_name":        apps.ClassName,
				"version_name":      apps.VersionName,
				"data_dir":          apps.DataDir,
				"is_user_installed": apps.UserInstall,
				"last_update_time":  apps.LastUpdateTime,
				"running_time":      apps.RinningTime,
				"last_time":         apps.LastTime,
			}, map[string]string{
				"name": apps.PackageName,
			})
			apps = &Androidapp{}
		}
	} else {
		acc.AddError(fmt.Errorf("this plugin only supported on android device"))
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
	VersionCode    string
	VersionName    string
	DataDir        string
	LabelName      string
	LastUpdateTime string
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
		case strings.Contains(line, "path:"):
			basepath := strings.Split(line, "path:")[1]
			if _, err := os.Stat("/data/hive/aapt"); err == nil {
				fmt.Println(strings.TrimSpace(basepath))
				getLable, _ := exec.Command("/data/hive/aapt", "dump", "badging", strings.TrimSpace(basepath)).Output()
				if strings.Contains(string(getLable), "application-label:") {
					{
						for _, linesofaapt := range strings.Split(string(getLable), "\n") {
							if strings.Contains(linesofaapt, "application-label:") {
								splitcolon := strings.Split(linesofaapt, ":")
								if len(splitcolon) > 0 {
									pkg.LabelName = strings.ReplaceAll(splitcolon[1], "'", "")
								}
							}
						}
					}
				}
			}
		case strings.Contains(line, "lastUpdateTime"):
			pkg.LastUpdateTime = strings.Split(line, "lastUpdateTime=")[1]
		}
	}
	return pkg
}

func DownloadFile(filepath string, url string) error {
	var err error
	resp, err := http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		out, err := os.Create(filepath)
		if err == nil {
			defer out.Close()
			io.Copy(out, resp.Body)
		}
	}
	return err
}

func GetUsageStats(appname string) (string, string) {
	var (
		usage    string
		lastTime string
	)
	getusage, err := exec.Command("dumpsys", "usagestats").Output()
	if err != nil {
		return "", ""
	}
	splitbyusage := strings.Split(string(getusage), "\n")
	for _, line := range splitbyusage {
		if strings.Contains(line, appname) && strings.Contains(line, "totalTime=") {
			scanner := bufio.NewScanner(strings.NewReader(line))
			scanner.Split(bufio.ScanWords)
			cols := make([]string, 0, 10)
			for scanner.Scan() {
				cols = append(cols, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return usage, lastTime
			}
			if len(cols) < 3 {
				break
			}
			re := regexp.MustCompile(`(([0-9])?[0-9]:)?[0-9][0-9]:[0-9][0-9]`)
			usage = re.FindString(cols[1])
			// lasttime
			lasttemp := cols[2] + " " + cols[3]
			reg := regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2})`)
			lastTime = reg.FindString(lasttemp)
			break
		}
	}
	return usage, lastTime
}
