package foreground_app

import (
	"fmt"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
display = :0
xauthority =: ~/.Xauthority
`

var Platform string

type ForegroundApp struct {
	Display    string
	Xauthority string
	Result     string
}

func (e *ForegroundApp) SampleConfig() string {
	return sampleConfig
}

func (fapps *ForegroundApp) Gather(acc telegraf.Accumulator) error {
	var app string
	var usage string
	var is_interactive bool
	Platform = GETPLATFORM()
	switch Platform {
	case "windows":
		if hwnd := GetWindow("GetForegroundWindow"); hwnd != 0 {
			app = GetWindowText(HWND(hwnd))
		}
	default:
		app, usage = GetFrontApp()
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
	}

	for k, v := range field {
		if reflect.ValueOf(v).IsZero() {
			delete(field, k)
		}
	}
	acc.AddFields("general_info", field, map[string]string{})

	fmt.Println(app, usage)
	return nil
}

func init() {
	inputs.Add("foreground_app", func() telegraf.Input {
		return &ForegroundApp{}
	})
}

func GETPLATFORM() string {
	OS_TYPE := runtime.GOOS
	if OS_TYPE == "linux" {
		execProp, err := exec.Command("getprop", "ro.product.board").Output()
		if err == nil {
			Platform := strings.TrimSuffix(string(execProp), "\n")
			if Platform != "" {
				OS_TYPE = "android"
			}
		}
	}
	return OS_TYPE
}
