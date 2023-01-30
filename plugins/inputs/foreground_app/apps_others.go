//go:build linux || darwin
// +build linux darwin

package foreground_app

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type (
	HANDLE uintptr
	HWND   HANDLE
)

func GetWindowText(hwnd HWND) string {
	return ""
}

func GetWindow(funcName string) uintptr {
	return 0
}

func GetFrontApp() (string, string) {
	var app string
	var usage string
	switch Platform {
	case "linux":
		SetEnvironment()
		app = GetForegoundApp()
	case "darwin":
		app = DarwinApp()
		usage = GetUsageStatsDarwin()
	case "android":
		app = AndroidForegroundApp()
		usage = GetUsageStats(app)
	}
	return app, usage
}

func AndroidForegroundApp() string {
	var (
		appname string
	)
	getappname, err := exec.Command("dumpsys", "window", "windows").Output()
	if err != nil {
		return appname
	}
	splitline := strings.Split(string(getappname), "\n")
	for _, line := range splitline {
		if strings.Contains(line, "mCurrentFocus=") {
			appnamesplit := strings.Split(line, "mCurrentFocus=")[1]
			appsplitslash := strings.Split(appnamesplit, "/")[0]
			splitspace := strings.Split(appsplitslash, " ")
			appname = splitspace[len(splitspace)-1]
			break
		}
	}
	return appname
}

func GetForegoundApp() string {
	var ForeApp string
	var Display_status bool
	if _, err := os.Stat("/tmp/.X11-unix"); os.IsNotExist(err) {
		return ""
	} else {
		dir, err := ioutil.ReadDir("/tmp/.X11-unix/")
		if err != nil {
			return ""
		}
		for _, d := range dir {
			if os.Getenv("DISPLAY") != "" {
				displayenv := strings.Replace(os.Getenv("DISPLAY"), ":", "X", -1)[0:2]
				if strings.Contains(d.Name(), displayenv) {
					Display_status = true
					break
				}
			}
		}
		if os.Getenv("DISPLAY") != "" && Display_status && os.Getenv("XAUTHORITY") != "" {
			var X *xgb.Conn
			X, err := xgb.NewConn()
			if err != nil {
				exec.Command("xhost", "si:localuser:root").Run()
				X, err = xgb.NewConn()
				if err != nil {
					return ForeApp
				}
			}
			if X != nil {
				setup := xproto.Setup(X)
				root := setup.DefaultScreen(X).Root
				aname := "_NET_ACTIVE_WINDOW"
				activeAtom, err := xproto.InternAtom(X, true, uint16(len(aname)),
					aname).Reply()
				if err != nil {
					return ""
				} else {
					aname = "_NET_WM_NAME"
					nameAtom, err := xproto.InternAtom(X, true, uint16(len(aname)),
						aname).Reply()
					if err != nil {
						return ""
					} else {
						reply, err := xproto.GetProperty(X, false, root, activeAtom.Atom,
							xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
						if err != nil {
							return ""
						} else {
							if len(reply.Value) > 0 {
								windowId := xproto.Window(xgb.Get32(reply.Value))
								reply, err = xproto.GetProperty(X, false, windowId, nameAtom.Atom,
									xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
								if err != nil {
									return ""
								} else {
									if reply.Value != nil {
										ForeApp = string(reply.Value)
									}
								}
							}
						}
					}
				}
			}
		} else {
			return ""
		}
	}
	return ForeApp
}

func DarwinApp() string {
	var (
		app string
	)
	frontapp, err := exec.Command("lsappinfo", "front").Output()
	if err != nil {
		return ""
	}
	appinfo, err := exec.Command("lsappinfo", "info", strings.TrimSpace(string(frontapp))).Output()
	if err != nil {
		return ""
	}
	spl := strings.Split(string(appinfo), "\n")
	for _, v := range spl {
		switch {
		case strings.Contains(v, strings.TrimSpace(string(frontapp))):
			re := regexp.MustCompile(`"([^"]*)"`)
			match := re.FindStringSubmatch(v)
			if len(match) > 0 {
				app = match[1]
			} else {
				return ""
			}
		}
	}
	return app
}

func GetUsageStatsDarwin() string {
	var (
		launch_time string
	)
	frontapp, err := exec.Command("lsappinfo", "front").Output()
	if err != nil {
		return ""
	}
	appinfo, err := exec.Command("lsappinfo", "info", strings.TrimSpace(string(frontapp))).Output()
	if err != nil {
		return ""
	}
	spl := strings.Split(string(appinfo), "\n")
	for _, v := range spl {
		switch {
		case strings.Contains(v, "launch time"):
			re := regexp.MustCompile(`(?P<year>\d{4})/(?P<month>\d{2})/(?P<day>\d{2})\s(?P<hour>\d{2}):(?P<minute>\d{2}):(?P<second>\d{2})`)
			match := re.FindStringSubmatch(v)
			if len(match) >= 0 {
				launch_time = match[0]
			} else {
				return ""
			}
		}
	}
	return launch_time
}
