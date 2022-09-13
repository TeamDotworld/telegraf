package mdm_agent

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type MDMAGENT struct {
	Version     int64 `toml:"version"`
	Name        string
	Command     string `toml:"command"`
	ProcessID   int    `toml:"process_id"`
	State       string
	Status      string
	NetworkMode string
	Time        int `toml:"time"`
}

func (e *MDMAGENT) SampleConfig() string {
	return sampleConfig
}

func (e *MDMAGENT) Description() string {
	return "MDM-Agent go-plugin for Telegraf"
}

func (e *MDMAGENT) Gather(acc telegraf.Accumulator) error {
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
	acc.AddFields("mdm_agent", map[string]interface{}{
		"name":         e.Name,
		"process_id":   e.ProcessID,
		"command":      e.Command,
		"state":        e.State,
		"network_mode": e.NetworkMode,
		"version":      e.Version,
		"status":       e.Status,
	}, map[string]string{
		"mdm-agent": "info",
	})
	return nil
}

func init() {
	inputs.Add("mdm_agent", func() telegraf.Input {
		return &MDMAGENT{}
	})
}

func getProcessUptime(pid int, sysUptimeSec int64) (int64, error) {
	// See glibc's /sysdeps/unix/sysv/linux/getclktck.c
	const SYSTEM_CLK_TCK = 100

	st, err := getProcessStartTime(pid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get starttime:", err)
		return 0, err
	}

	procUptime := sysUptimeSec - (int64(st) / SYSTEM_CLK_TCK)
	return procUptime, nil
}

func getProcessStartTime(pid int) (int64, error) {
	// Index of the starttime field. See proc(5).
	const StartTimeIndex = 21

	fname := filepath.Join("/proc", strconv.Itoa(pid), "stat")
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return 0, err
	}

	fields := bytes.Fields(data)
	if len(fields) < StartTimeIndex+1 {
		return 0, fmt.Errorf("invalid /proc/[pid]/stat format: too few fields: %d", len(fields))
	}

	s := string(fields[StartTimeIndex])
	starttime, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid starttime: %d", err)
	}

	return starttime, nil
}
