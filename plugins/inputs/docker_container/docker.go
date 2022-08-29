package docker_container

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	docclient "github.com/docker/docker/client"
	"github.com/influxdata/telegraf"
	dockerstats "github.com/influxdata/telegraf/internal/docker_stats"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type DockerContainer struct {
}

func (e *DockerContainer) SampleConfig() string {
	return sampleConfig
}

func (e *DockerContainer) Description() string {
	return "Getdocker container go-plugin for Telegraf"
}

func (e *DockerContainer) Gather(acc telegraf.Accumulator) error {
	current_stats, err := dockerstats.Current()
	if err != nil {
		acc.AddError(err)
	}
	cli, err := docclient.NewClientWithOpts(docclient.FromEnv)
	if err != nil {
		acc.AddError(err)
	}
	containerss, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		acc.AddError(err)
	}
	if len(current_stats) > 0 {
		for _, s := range current_stats {
			CPU, _ := strconv.ParseFloat(strings.TrimSuffix(s.CPU, "%"), 64)
			docmem := strings.Split(s.Memory.Raw, " / ")
			letteri0, numberi0 := ParseFlight(docmem[0])
			letteri1, numberi1 := ParseFlight(docmem[1])
			MEMUsage := ConvertB(letteri0, numberi0)
			MEMLimit := ConvertB(letteri1, numberi1)
			MEMPercentage, _ := strconv.ParseFloat(strings.TrimSuffix(s.Memory.Percent, "%"), 64)
			docnet := strings.Split(s.IO.Network, " / ")
			NETReceiver := docnet[0]
			NETSend := docnet[1]
			docblock := strings.Split(s.IO.Block, " / ")
			BlockWriten := docblock[0]
			BlockRead := docblock[1]
			var (
				Status                string
				Command               string
				Created               int64
				HostConfigNetworkMode string
				State                 string
				Mounts                string
				ImageName             string
				Version               string
			)
			for _, contmounts := range containerss {
				if s.Container == contmounts.ID[:12] {
					Status = contmounts.Status
					Command = contmounts.Command
					Created = contmounts.Created
					HostConfigNetworkMode = contmounts.HostConfig.NetworkMode
					State = contmounts.State
					Mounts = GetContainerMounts(contmounts)
					if strings.Contains(contmounts.Image, ":") {
						ver := strings.Split(contmounts.Image, ":")
						if len(ver[1]) < 10 {
							ImageName = ver[0]
							Version = ver[1]
						} else {
							ImageName = ver[1]
							Version = "latest"
						}
					} else {
						ImageName = contmounts.Image
						Version = "latest"
					}
				}
			}
			acc.AddFields("docker_container", map[string]interface{}{
				"container":      s.Container,
				"name":           s.Name,
				"cpu":            CPU,
				"mem_usage":      MEMUsage,
				"mem_limit":      MEMLimit,
				"mem_percentage": MEMPercentage,
				"net_receiver":   NETReceiver,
				"net_sender":     NETSend,
				"block_writen":   BlockWriten,
				"block_read":     BlockRead,
				"status":         Status,
				"command":        Command,
				"created":        Created,
				"network_mode":   HostConfigNetworkMode,
				"state":          State,
				"data_dir":       Mounts,
				"package_name":   ImageName,
				"version":        Version,
			}, map[string]string{
				"name": s.Name,
			})
		}
	}
	return nil
}

func init() {
	inputs.Add("docker_container", func() telegraf.Input {
		return &DockerContainer{}
	})
}
func GetContainerMounts(container types.Container) string {
	var mounts []string
	var seporator = ", "
	for _, m := range container.Mounts {
		mount := fmt.Sprintf("%v:%v", m.Source, m.Destination)
		mounts = append(mounts, mount)
	}
	return strings.Join(mounts, seporator)
}

func ParseFlight(s string) (letters, numbers string) {
	var l, n []rune
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			l = append(l, r)
		case r >= 'a' && r <= 'z':
			l = append(l, r)
		case r >= '0' && r <= '9':
			n = append(n, r)
		case r >= '.' && r <= '9':
			n = append(n, r)
		}
	}
	return string(l), string(n)
}

func ConvertB(letter string, number string) float64 {
	var ByteNumber float64
	ConvertInt, _ := strconv.ParseFloat(number, 64)
	if letter == "GiB" {
		ByteNumber = ConvertInt * 1024 * 1024 * 1024
	} else if letter == "MiB" {
		ByteNumber = ConvertInt * 1024 * 1024
	} else {
		ByteNumber = ConvertInt
	}
	return ByteNumber
}
