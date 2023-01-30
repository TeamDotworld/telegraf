//go:build linux || windows
// +build linux windows

package cpu_temp

func GetCPUtemp() string {
	return ""
}
