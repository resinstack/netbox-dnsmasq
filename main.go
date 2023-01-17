package main

import (
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/resinstack/netbox-dnsmasq/netbox"
)

// DHCPHost contains the information required to generate a matching
// line in a dhcp configuration daemon.
type DHCPHost struct {
	DeviceID int64
	HWAddr   []string
	Name     string
	Addr     string
}

func main() {
	if _, verbose := os.LookupEnv("VERBOSE"); !verbose {
		log.SetOutput(io.Discard)
	}

	hostTmplStr := "{{JoinStrings .HWAddr \",\"}},{{.Addr}}\n"
	if hts := os.Getenv("DNSMASQ_TEMPLATE"); hts != "" {
		hostTmplStr = hts
	}
	hostTmpl := template.New("dhcp-host")
	hostTmpl = hostTmpl.Funcs(template.FuncMap{"JoinStrings": strings.Join})
	hostTmpl, err := hostTmpl.Parse(hostTmplStr)
	if err != nil {
		log.Println("Error parsing dhcp-host template", err)
		os.Exit(1)
	}

	token := os.Getenv("NETBOX_TOKEN")
	if token == "" {
		log.Println("Please provide netbox API token via env var NETBOX_TOKEN")
		os.Exit(1)
	}

	netboxURL := os.Getenv("NETBOX_URL")
	if netboxURL == "" {
		log.Println("Please provide netbox url via NETBOX_URL")
		os.Exit(1)
	}

	site := os.Getenv("NETBOX_SITE")
	nb, err := netbox.NewClient(netbox.WithNetBoxURL(netboxURL), netbox.WithToken(token))
	if err != nil {
		log.Println("Error initializing client:", err)
		os.Exit(1)
	}

	devices, err := nb.ListDevices(site)
	if err != nil {
		log.Println("Error listing devices:", err)
		os.Exit(1)
	}

	hosts := make(map[int64]*DHCPHost, len(devices))
	for _, dev := range devices {
		ipaddr := strings.SplitN(dev.PrimaryIPv4.Address, "/", 2)[0]

		ifres, err := nb.ListInterfaces(dev.ID)
		if err != nil {
			log.Printf("Error pulling interfaces for %s: %v", dev.Name, err)
			continue
		}
		if len(ifres) == 0 {
			log.Printf("No interface available for PXE! (%s)", dev.Name)
			continue
		}

		hwaddrs := make([]string, len(ifres))
		for i, int := range ifres {
			if int.MACAddress == "" {
				continue
			}
			hwaddrs[i] = strings.ToLower(int.MACAddress)
		}

		log.Println(dev.ID, dev.Name, ipaddr, hwaddrs)
		hosts[dev.ID] = &DHCPHost{
			DeviceID: dev.ID,
			Name:     dev.Name,
			Addr:     ipaddr,
			HWAddr:   hwaddrs,
		}
	}

	for _, host := range hosts {
		if err := hostTmpl.Execute(os.Stdout, host); err != nil {
			log.Println("Error executing template", err)
		}
	}
}
