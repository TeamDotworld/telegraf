package geo_lookup

import (
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type GeoLookup struct {
	Latitide  string
	Langitude string
}

func (e *GeoLookup) SampleConfig() string {
	return sampleConfig
}

func (e *GeoLookup) Description() string {
	return "Get Exact location of your system"
}

func (e *GeoLookup) Gather(acc telegraf.Accumulator) error {
	platform := GETPLATFORM()
	switch platform {
	case "android":
		location, err := exec.Command("dumpsys", "location").Output()
		if err != nil {
			return err
		} else {
			splibyline := strings.Split(string(location), "\n")
			for _, line := range splibyline {
				if strings.Contains(line, "gps ") {
					re := regexp.MustCompile(`[+-]?([0-9]+([.][0-9]*)?|[.][0-9]+)`)
					match := re.FindStringSubmatch(line)
					if len(match) > 0 {
						e.Latitide = match[0]
						e.Langitude = match[1]
						break
					}
				}
			}
		}
	}

	acc.AddFields("geo_lookup", map[string]interface{}{
		"latitude":  e.Langitude,
		"langitude": e.Latitide,
	}, map[string]string{})
	return nil
}

func init() {
	inputs.Add("geo_lookup", func() telegraf.Input {
		return &GeoLookup{}
	})
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
