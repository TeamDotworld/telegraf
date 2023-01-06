package win_apps

import (
	"encoding/json"
	"log"
	"os/exec"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type Apps struct {
	Name string
}

func (a *Apps) SampleConfig() string {
	return sampleConfig
}

func (a *Apps) Description() string {
	return "Win Apps go-plugin for Telegraf"
}

func (a *Apps) Gather(acc telegraf.Accumulator) error {
	args := "Get-AppxPackage | ConvertTo-Json"
	appslist, err := exec.Command("powershell.exe", args).Output()
	if err != nil {
		log.Println("Failed to get apps")
		return err
	}
	var apps []map[string]interface{}
	err = json.Unmarshal(appslist, &apps)
	if err != nil {
		log.Println("Failed to unmarshal apps")
		return err
	}
	if len(apps) > 0 {
		for _, v := range apps {
			acc.AddFields("win_apps", v, map[string]string{
				"name": v["Name"].(string),
			})
		}
	}

	return nil
}

func init() {
	inputs.Add("win_apps", func() telegraf.Input {
		return &Apps{}
	})
}
