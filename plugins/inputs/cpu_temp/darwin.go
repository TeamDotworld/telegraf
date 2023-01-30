//go:build darwin
// +build darwin

package cpu_temp

import (
	"golang.org/x/sys/unix"
)

func GetCPUtemp() string {
	thermal, err := unix.Sysctl("machdep.xcpm.cpu_thermal_level")
	if err != nil {
		return ""
	}
	return thermal
}
