package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sort"
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

// ShoelacesHost is a host that can be mapped without user input in
// shoelaces.
type ShoelacesHost struct {
	Network string     `json:"network"`
	Script  ShoeScript `json:"script"`
}

// ShoeScript is the nested script type for shoelaces.  It is only
// supported to specify by name.
type ShoeScript struct {
	Name string `json:"name"`
}

// ShoelacesNetworkMap matches the structure that shoelaces expects to
// be able to read in.
type ShoelacesNetworkMap struct {
	NetworkMaps []ShoelacesHost `json:"networkMaps"`
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

	nb, err := netbox.NewClient(netbox.WithNetBoxURL(netboxURL), netbox.WithToken(token))
	if err != nil {
		log.Println("Error initializing client:", err)
		os.Exit(1)
	}

	site := os.Getenv("NETBOX_SITE")
	tag := os.Getenv("NETBOX_TAG")
	shoetag := os.Getenv("SHOELACES_TAG_PREFIX")
	devices, err := nb.ListDevices(site, tag)
	if err != nil {
		log.Println("Error listing devices:", err)
		os.Exit(1)
	}

	hosts := make(map[int64]*DHCPHost, len(devices))
	shoenets := []ShoelacesHost{}
	for _, dev := range devices {
		if dev.PrimaryIPv4.Address == "" {
			continue
		}
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

		// Construct the shoelaces mapping if enabled.
		if shoetag != "" {
			for _, tag := range dev.Tags {
				if strings.HasPrefix(tag.Slug, shoetag) {
					// Map this script for this host's IPs
					shoehost := ShoelacesHost{
						Network: ipaddr + "/32",
						Script: ShoeScript{
							Name: strings.TrimPrefix(tag.Slug, shoetag) + ".ipxe",
						},
					}
					shoenets = append(shoenets, shoehost)
					break
				}
			}
		}
	}

	if os.Getenv("DNSMASQ_HOSTSFILE") != "" {
		tHosts := make([]*DHCPHost, len(hosts))
		i := 0
		for _, host := range hosts {
			sort.Strings(host.HWAddr)
			tHosts[i] = host
			i++
		}
		sort.Slice(tHosts, func(i, j int) bool {
			return tHosts[i].HWAddr[0] < tHosts[j].HWAddr[0]
		})

		f, err := os.Create(os.Getenv("DNSMASQ_HOSTSFILE"))
		if err != nil {
			log.Println("Error writing out hosts file", err)
			os.Exit(1)
		}
		defer f.Close()

		for _, host := range tHosts {
			if err := hostTmpl.Execute(f, host); err != nil {
				log.Println("Error executing template", err)
			}
		}
	}

	if os.Getenv("SHOELACES_MAPFILE") != "" {
		f, err := os.Create(os.Getenv("SHOELACES_MAPFILE"))
		if err != nil {
			log.Println("Error opening shoelaces map file", err)
			os.Exit(1)
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		if err := enc.Encode(ShoelacesNetworkMap{NetworkMaps: shoenets}); err != nil {
			log.Println("Error writing shoelaces mappings", err)
			os.Exit(1)
		}
	}
}
