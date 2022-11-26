package foreground_app

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
display = :0
xauthority =: ~/.Xauthority
`

type ForegroundApp struct {
	Display    string
	Xauthority string
	Result     string
}

func (e *ForegroundApp) SampleConfig() string {
	return sampleConfig
}

func (e *ForegroundApp) Description() string {
	return "Get Foreground application based on platform using go-plugin for Telegraf"
}

func (fapps *ForegroundApp) Gather(acc telegraf.Accumulator) error {
	var (
		app            string
		is_interactive bool
		// blestate       bool
		usage string
	)
	platform := GETPLATFORM()
	if platform == "linux" {
		SetEnvironment()
		app = GetLinuxForegoundApp()
		// blestate = LinuxBleStatus()
	} else if platform == "android" {
		app = AndroidForegroundApp()
		usage = GetUsageStats(app)
		// blestate = AndroidBleState()
	} else if platform == "windows" {
		if hwnd := GetWindow("GetForegroundWindow"); hwnd != 0 {
			app = GetWindowText(HWND(hwnd))
		} else {
			app = "none"
		}
	}
	if app != "" {
		is_interactive = true
	} else {
		is_interactive = false
	}

	field := map[string]interface{}{
		"is_interactive":                 is_interactive,
		"Current_foreground_application": app,
		"logged":                         time.Now().Format(time.RFC3339),
		"running_time":                   usage,
		// "ble_state":                      blestate,
	}
	for k, v := range field {
		if reflect.ValueOf(v).IsZero() {
			delete(field, k)
		}
	}

	acc.AddFields("general_info", field, map[string]string{})
	return nil
}

func LinuxBleStatus() bool {
	cmd, err := exec.Command("hcitool", "dev").Output()
	if err == nil {
		reg_mac := regexp.MustCompile(`[0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}`)
		if reg_mac.MatchString(string(cmd)) {
			return true
		}
	}
	return false
}

func AndroidBleState() bool {
	cmd, err := exec.Command("settings", "list", "global").Output()
	if err == nil {
		splitline := strings.Split(string(cmd), " ")
		for _, line := range splitline {
			if strings.Contains(line, "bluetooth_on") {
				ref := regexp.MustCompile(`[1]`)
				if ref.MatchString(string(cmd)) {
					return true
				}
			}
		}
	}
	return false
}

func WindowsBleState() bool {
	cmd, err := exec.Command("sc", "query", "bthserv").Output()
	if err == nil {
		if strings.Contains(string(cmd), "RUNNING") {
			return true
		}
	}
	return false
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
func SetEnvironment() {
	if runtime.GOOS == "linux" {
		MonitorNumber := GetMonitors()
		if len(MonitorNumber) > 0 {
			os.Setenv("DISPLAY", ":"+strconv.Itoa(MonitorNumber[0]))
		}
		if len(MonitorNumber) > 0 {
			if MonitorNumber[0] == 5 {
				os.Setenv("XAUTHORITY", "/home/dwmdm/.Xauthority")
			} else {
				homeuser := GetHomeUsers()
				os.Setenv("XAUTHORITY", "/home/"+homeuser+"/.Xauthority")
			}
		}
	}
}
func GetHomeUsers() string {
	var userName string
	execHome, err := exec.Command("ls", "/home").Output()
	if err != nil {
		return ""
	}
	split := strings.Split(string(execHome), "\n")
	for _, name := range split {
		if !strings.Contains(name, "dwmdm") {
			if GetUser(name) {
				userName = name
				break
			}
		}
	}
	return userName
}
func GetUser(userCheck string) bool {
	var usernameUser bool
	var Users []string
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return false
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		// skip all line starting with #
		if equal := strings.Index(line, "#"); equal < 0 {
			// get the username and description
			lineSlice := strings.FieldsFunc(line, func(divide rune) bool {
				return divide == ':' // we divide at colon
			})

			if len(lineSlice) > 0 {
				Users = append(Users, lineSlice[0])
			}

		}

		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}
	for _, name := range Users {
		if strings.Contains(name, userCheck) {
			usernameUser = true
			break
		} else {
			usernameUser = false
		}
	}
	return usernameUser
}
func GetMonitors() []int {
	var MonitorNumber []int
	if _, err := os.Stat("/tmp/.X11-unix"); err == nil {
		xradr, err := exec.Command("ls", "/tmp/.X11-unix").Output()
		if err != nil {
			return MonitorNumber
		}
		xradrs := strings.ReplaceAll(string(xradr), "\n", "")
		xrandr := strings.ReplaceAll(xradrs, "X", ":")
		totaldisplay := strings.Split(xrandr, ":")
		for _, v := range totaldisplay {
			if v != "" {
				convInt, err := strconv.Atoi(v)
				if err != nil {
					return MonitorNumber
				}
				MonitorNumber = append(MonitorNumber, convInt)
			}
		}
		sort.Ints(MonitorNumber)
	}
	return MonitorNumber
}

func init() {
	inputs.Add("foreground_app", func() telegraf.Input {
		return &ForegroundApp{}
	})
}

func GetUsageStats(appname string) string {
	var usage string
	getusage, err := exec.Command("dumpsys", "usagestats").Output()
	if err != nil {
		return ""
	}
	splitbyusage := strings.Split(string(getusage), "\n")
	for _, line := range splitbyusage {
		if strings.Contains(line, appname) && strings.Contains(line, "totalTime=") {
			re := regexp.MustCompile(`[0-9][0-9]:[0-9][0-9]`)
			usage = re.FindString(line)
			break
		}
	}
	return usage
}
