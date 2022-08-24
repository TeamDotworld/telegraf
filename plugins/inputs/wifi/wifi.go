package wifi

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string
var OS_TYPE string

type WiFi struct {
	WifiName        string
	WifiMAC         string
	NetworkID       string
	ConnectedWifiIP string
	BSSID           string
	LinkSpeed       string
}

func (wifi *WiFi) SampleConfig() string {
	return sampleConfig
}
func (wifi *WiFi) Description() string {
	return "WiFi go-plugin for Telegraf"
}
func (wifi *WiFi) Gather(acc telegraf.Accumulator) error {
	platform := GETPLATFORM()
	if platform == "android" {
		getConnectedWifi, err := exec.Command("dumpsys", "wifi").Output()
		if err != nil {
			return err
		}
		splitline := strings.Split(string(getConnectedWifi), "\n")
		var WifiInfo string
		for _, line := range splitline {
			if strings.Contains(line, "WifiInfo") {
				WifiInfo = line
				break
			}
		}
		if WifiInfo != "" {
			WifiInfo = strings.TrimSpace(WifiInfo)
			splitline := strings.Split(WifiInfo, ",")
			for _, line := range splitline {
				switch {
				case strings.Contains(line, "SSID:") && !strings.Contains(line, "BSSID:"):
					wifi.WifiName = strings.Split(line, ":")[1]
					wifi.WifiName = strings.TrimSpace(wifi.WifiName)
				case strings.Contains(line, "BSSID:"):
					wifi.BSSID = strings.Split(line, "BSSID:")[1]
					wifi.BSSID = strings.TrimSpace(wifi.BSSID)
				case strings.Contains(line, "Link"):
					re := regexp.MustCompile(`\d+`)
					wifi.LinkSpeed = re.FindString(line)
				case strings.Contains(line, "Net ID"):
					wifi.NetworkID = strings.Split(line, ":")[1]
					wifi.NetworkID = strings.TrimSpace(wifi.NetworkID)
				case strings.Contains(line, "MAC:"):
					wifi.WifiMAC = strings.Split(line, "MAC:")[1]
					wifi.WifiMAC = strings.TrimSpace(wifi.WifiMAC)
				}
			}
		}
	} else if platform == "linux" {
		wname, _ := exec.Command("iwgetid", "--raw").Output()
		wifi.WifiName = strings.TrimSuffix(string(wname), "\n")
		iface := GetInterface()
		ifas, _ := net.Interfaces()
		for _, ifa := range ifas {
			if ifa.Name == iface {
				wifi.WifiMAC = ifa.HardwareAddr.String()
			}
		}
		wifi.NetworkID = ReadTxtFile("/sys/class/net/" + iface + "/netdev_group")
		// Access Point MAC Address
		bssid, err := exec.Command("iwgetid", "-a").Output()
		if err != nil {
			wifi.BSSID = ""
		} else {
			reg_mac := regexp.MustCompile(`[0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}`)
			wifi.BSSID = reg_mac.FindString(string(bssid))
		}
		iwconfig, err := exec.Command("iwconfig", iface).Output()
		if err != nil {
			fmt.Println("Error of get iwconfig: ", err)
		}
		// reg for Signal level=
		reg := "Signal level=-?[0-9]+"
		mustc := regexp.MustCompile(reg)
		match := mustc.FindAllString(string(iwconfig), -1)
		if len(match) > 0 {
			wifi.LinkSpeed = match[0][len("Signal level="):]
		}
	} else if platform == "windows" {
		getinterfaces, err := exec.Command("netsh", "wlan", "show", "interfaces").Output()
		if err != nil {
			fmt.Println(err)
		}
		getinterfacesString := string(getinterfaces)
		getinterfacesString = strings.TrimSpace(getinterfacesString)
		splitinterfaces := strings.Split(getinterfacesString, "\n")
		for _, value := range splitinterfaces {
			if strings.Contains(value, "SSID") && !strings.Contains(value, "BSSID") {
				wifi.WifiName = strings.TrimSpace(strings.Split(value, ":")[1])
			}
			if strings.Contains(value, "BSSID") {
				reg_mac := regexp.MustCompile(`[0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}`)
				wifi.BSSID = reg_mac.FindString(value)
			}
			if strings.Contains(value, "Channel") {
				wifi.NetworkID = strings.TrimSpace(strings.Split(value, ":")[1])
			}
			if strings.Contains(value, "Physical address") {
				reg_mac := regexp.MustCompile(`[0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}`)
				wifi.WifiMAC = reg_mac.FindString(value)
			}
			if strings.Contains(value, "Transmit rate") {
				reg_int := regexp.MustCompile(`[0-9]{1,}`)
				wifi.LinkSpeed = reg_int.FindString(value)
			}
		}
	}
	wifi.ConnectedWifiIP = GetOutboundIP().String()
	acc.AddFields("wifi", map[string]interface{}{
		"wifi_name":         wifi.WifiName,
		"bssid":             wifi.BSSID,
		"link_speed":        wifi.LinkSpeed,
		"network_id":        wifi.NetworkID,
		"wifi_mac":          wifi.WifiMAC,
		"connected_wifi_ip": wifi.ConnectedWifiIP,
	}, map[string]string{})
	return nil
}
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
func GETPLATFORM() string {
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

func GetInterface() string {
	intf, err := NetworkInterfaces()
	if err != nil {
		fmt.Println(err)
	}

	var interface_name string
	if len(intf) > 0 {
		interface_name = string(intf[0])
	} else {
		interface_name = "eth0"
	}
	return interface_name
}

type NetworkInterface string

func NetworkInterfaces() ([]NetworkInterface, error) {
	b, err := ioutil.ReadFile("/proc/net/wireless")
	if err != nil {
		return nil, err
	}

	var interfaces []NetworkInterface
	for _, line := range strings.Split(string(b), "\n")[2:] {
		parts := strings.Split(line, ":")
		i := strings.TrimSpace(parts[0])
		if i != "" {
			interfaces = append(interfaces, NetworkInterface(i))
		}
	}
	return interfaces, nil
}

func init() {
	inputs.Add("wifi", func() telegraf.Input {
		return &WiFi{}
	})
}

func ReadTxtFile(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	var content string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += strings.TrimSpace(scanner.Text())
	}
	return content
}
