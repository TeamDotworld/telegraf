//go:build darwin
// +build darwin

package battery

func GetBattery(platform string) Battery {
	return Battery{}
}
