package foreground_app

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func SetEnvironment() {
	if runtime.GOOS == "linux" {
		MonitorNumber := GetMonitors()
		if len(MonitorNumber) > 0 {
			os.Setenv("DISPLAY", MonitorNumber[0])
			if MonitorNumber[0] == "5" {
				os.Setenv("XAUTHORITY", "/home/dothive/.Xauthority")
			} else {
				homeuser := GetCurrentHomeUser()
				os.Setenv("XAUTHORITY", "/home/"+homeuser+"/.Xauthority")
			}
		}
	}
}

func GetCurrentHomeUser() string {
	var userName string
	getCurrentUser, _, err := RunCommand("w", "", "-s", "-h")
	if err != nil {
		return ""
	}
	spl := strings.Split(getCurrentUser, "\n")
	if len(spl) > 0 {
		splspace := strings.Split(spl[0], " ")
		userName = strings.TrimSpace(splspace[0])
	}
	return userName
}
func RunCommand(command string, dir string, args ...string) (string, string, error) {
	// Create buffers for stdout and stderr
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	// Create command
	cmd := exec.Command(command, args...)
	// Set stdout and stderr
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Set directory if given
	if dir != "" {
		cmd.Dir = dir
	}
	// Run command
	err := cmd.Run()
	if err != nil || stderr.String() != "" {
		return stdout.String(), stderr.String(), err
	}
	// Return stdout, stderr, and error
	return stdout.String(), stderr.String(), err
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

func GetMonitors() []string {
	var MonitorNumber []string
	if _, err := os.Stat("/tmp/.X11-unix"); err == nil {
		dir, err := os.ReadDir("/tmp/.X11-unix")
		if err != nil {
			return MonitorNumber
		}
		for _, file := range dir {
			monitorcnt := strings.ReplaceAll(file.Name(), "X", ":")
			MonitorNumber = append(MonitorNumber, monitorcnt)
		}
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
