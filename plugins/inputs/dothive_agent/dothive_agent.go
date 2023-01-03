package dothive_agent

import (
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type DOTHIVEAGENT struct {
	Version     int64 `toml:"version"`
	Name        string
	Command     string `toml:"command"`
	ProcessID   int    `toml:"process_id"`
	State       string
	Status      string
	NetworkMode string
	Time        int `toml:"time"`
}

func (e *DOTHIVEAGENT) SampleConfig() string {
	return sampleConfig
}

func (e *DOTHIVEAGENT) Description() string {
	return "Dothive-Agent go-plugin for Telegraf"
}

func (e *DOTHIVEAGENT) Gather(acc telegraf.Accumulator) error {
	if e.Command != "" {
		if strings.Contains(e.Command, "/") {
			str := strings.Split(e.Command, "/")
			e.Name = str[len(str)-1]
		} else if strings.Contains(e.Command, "\\") {
			str := strings.Split(e.Command, "\\")
			e.Name = str[len(str)-1]
		} else {
			e.Name = e.Command
		}
		e.Time = int(time.Now().Second()) + e.Time
		if e.Time > 60 {
			uptimeMin := e.Time / 60
			if uptimeMin > 60 {
				uptimeHour := uptimeMin / 60
				e.Status = strconv.Itoa(uptimeHour) + " Hours"
			} else if uptimeMin < 60 {
				e.Status = strconv.Itoa(uptimeMin) + " Minutes"
			}
		} else if e.Time < 60 {
			e.Status = strconv.Itoa(e.Time) + " seconds"
		}
	}
	if e.ProcessID != 0 {
		e.State = "running"
		e.NetworkMode = "default"
	}
	acc.AddFields("dothive_agent", map[string]interface{}{
		"name":         e.Name,
		"process_id":   e.ProcessID,
		"command":      e.Command,
		"state":        e.State,
		"network_mode": e.NetworkMode,
		"version":      e.Version,
		"status":       e.Status,
	}, map[string]string{
		"dothive-agent": "info",
	})
	return nil
}

func init() {
	inputs.Add("dothive_agent", func() telegraf.Input {
		return &DOTHIVEAGENT{}
	})
}
