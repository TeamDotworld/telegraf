package geoip

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

type GeoIP struct {
	Ip       string
	City     string
	Region   string
	Country  string
	Location string
	Org      string
	Postal   string
}

func (e *GeoIP) SampleConfig() string {
	return sampleConfig
}

func (e *GeoIP) Description() string {
	return "Get Exact location of your system"
}

func (e *GeoIP) Gather(acc telegraf.Accumulator) error {
	var IPAddress string
	req, err := http.NewRequest("GET", "https://ipinfo.io/ip", nil)
	if err == nil {
		client := &http.Client{}
		resp, err1 := client.Do(req)
		if err1 == nil {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				IPAddress = strings.TrimSuffix(strings.TrimSpace(string(body)), "\n")
			}
		}
	}
	if IPAddress != "" {
		reqs, err := http.NewRequest("GET", "https://ipinfo.io/"+IPAddress, nil)
		if err == nil {
			clients := &http.Client{}
			resp, err := clients.Do(reqs)
			if err == nil {
				if resp.StatusCode == 200 {
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err == nil {
						_ = json.Unmarshal(body, &e)
					}
				}
			}
		}
	}
	acc.AddFields("geoip", map[string]interface{}{
		"public_ip": e.Ip,
		"city":      e.City,
		"region":    e.Region,
		"country":   e.Country,
		"location":  e.Location,
		"org":       e.Org,
		"postal":    e.Postal,
	}, map[string]string{})
	return nil
}

func init() {
	inputs.Add("geoip", func() telegraf.Input {
		return &GeoIP{}
	})
}
