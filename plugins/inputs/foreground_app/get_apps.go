//go:build linux || darwin
// +build linux darwin

package foreground_app

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func GetLinuxForegoundApp() string {
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
