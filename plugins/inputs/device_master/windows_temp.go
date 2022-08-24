//go:build linux || darwin
// +build linux darwin

package device_master

func GetWindowsMetaData() DeviceMaster {
	return DeviceMaster{}
}
