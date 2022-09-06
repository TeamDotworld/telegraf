package ip_uiniverse

import (
	"net"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type IP_Universe struct {
	IpAddress string
	Network   string
}

func (e *IP_Universe) SampleConfig() string {
	return sampleConfig
}

func (e *IP_Universe) Description() string {
	return "Get all system ip address go-plugin for Telegraf"
}

func (ip *IP_Universe) Gather(acc telegraf.Accumulator) error {
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err == nil {
				for _, a := range addrs {
					var interface_type string
					v4_add := a.(*net.IPNet).IP.To4()
					if v4_add != nil {
						interface_type = "v4"
					} else {
						interface_type = "v6"
					}
					acc.AddFields("ip_address", map[string]interface{}{
						"hardware_address": i.HardwareAddr.String(),
						"flag":             i.Flags.String(),
						"ip":               a.String(),
						"mtu":              i.MTU,
					}, map[string]string{
						"interface_type": i.Name + interface_type,
					})
				}
			}
		}
	}
	return nil
}

func init() {
	inputs.Add("ip_universe", func() telegraf.Input {
		return &IP_Universe{}
	})
}
