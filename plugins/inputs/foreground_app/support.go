package foreground_app

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func SetEnvironment() {
	if runtime.GOOS == "linux" {
		MonitorNumber := GetMonitors()
		if len(MonitorNumber) > 0 {
			os.Setenv("DISPLAY", ":"+strconv.Itoa(MonitorNumber[0]))
			if MonitorNumber[0] == 5 {
				os.Setenv("XAUTHORITY", "/home/dothive/.Xauthority")
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
			if CheckUser(name) {
				userName = name
				break
			}
		}
	}
	return userName
}

func CheckUser(userCheck string) bool {
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

func GetUsageStats(appname string) string {
	var usage string
	getusage, err := exec.Command("dumpsys", "usagestats").Output()
	if err != nil {
		return ""
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
				return usage
			}
			re := regexp.MustCompile(`(([0-9])?[0-9]:)?[0-9][0-9]:[0-9][0-9]`)
			usage = re.FindString(cols[1])
			break
		}
	}
	return usage
}
