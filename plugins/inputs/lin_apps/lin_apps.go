package lin_apps

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type Service struct {
	Names        string
	User         string
	Group        string
	FragmentPath string
	Descriptions string
	SubState     string
	Type         string
	MainPID      string
}

func (s Service) SampleConfig() string {
	return sampleConfig
}
func (s Service) Description() string {
	return "Linux service go-plugin for Telegraf"
}

func (s Service) Gather(acc telegraf.Accumulator) error {
	if GETPLATFORM() == "linux" {
		list_services := []string{}
		files, err := os.ReadDir("/etc/systemd/system/")
		if err != nil {
			acc.AddError(fmt.Errorf("systemd is not found"))
			return err
		}

		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".service") && !strings.HasPrefix(f.Name(), "dothive") {
				list_services = append(list_services, f.Name())
			}
		}
		if len(list_services) == 0 {
			acc.AddError(fmt.Errorf("no such services in a system"))
			return err
		}

		for _, service := range list_services {
			s = GetStatsOfService(service)
			acc.AddFields("lin_apps", map[string]interface{}{
				"names":         s.Names,
				"user":          s.User,
				"group":         s.Group,
				"fragment_path": s.FragmentPath,
				"descriptions":  s.Descriptions,
				"sub_state":     s.SubState,
				"type":          s.Type,
				"main_pid":      s.MainPID,
			}, map[string]string{
				"names": s.Names,
			})
		}
	} else {
		acc.AddError(fmt.Errorf("this plugin only supported on linux device"))
	}
	return nil
}
func init() {
	inputs.Add("lin_apps", func() telegraf.Input {
		return &Service{}
	})
}

func GetStatsOfService(servicename string) Service {
	var service Service
	show, err := exec.Command("systemctl", "show", servicename).Output()
	if err != nil {
		fmt.Println(err)
		return Service{}
	}
	getbyline := strings.Split(string(show), "\n")
	for _, line := range getbyline {
		switch true {
		case strings.HasPrefix(line, "Names"):
			service.Names = GetValueOfVariable(line)
		case strings.HasPrefix(line, "User"):
			service.User = GetValueOfVariable(line)
		case strings.HasPrefix(line, "Group"):
			service.Group = GetValueOfVariable(line)
		case strings.HasPrefix(line, "FragmentPath"):
			service.FragmentPath = GetValueOfVariable(line)
		case strings.HasPrefix(line, "Description"):
			service.Descriptions = GetValueOfVariable(line)
		case strings.HasPrefix(line, "SubState"):
			service.SubState = GetValueOfVariable(line)
		case strings.HasPrefix(line, "Type"):
			service.Type = GetValueOfVariable(line)
		case strings.HasPrefix(line, "MainPID"):
			service.MainPID = GetValueOfVariable(line)
		}
	}
	return service
}

func GetValueOfVariable(line string) string {
	splitbyequal := strings.Split(line, "=")
	return splitbyequal[1]
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
