package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
)

type DHCPHost struct {
	DeviceID int64
	HWAddr   []string
	Name     string
	Addr     string
}

func strPtr(s string) *string {
	return &s
}

func main() {
	if _, verbose := os.LookupEnv("VERBOSE"); !verbose {
		log.SetOutput(io.Discard)
	}

	site := os.Getenv("NETBOX_SITE")

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

	netboxHost := os.Getenv("NETBOX_HOST")
	if netboxHost == "" {
		log.Println("Please provide netbox host via env var NETBOX_HOST")
		os.Exit(1)
	}

	protocol := os.Getenv("NETBOX_PROTOCOL")
	if protocol == "" {
		protocol = "https"
	}

	transport := httptransport.New(netboxHost, client.DefaultBasePath, []string{protocol})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)

	c := client.New(transport, nil)
	req := dcim.NewDcimDevicesListParams()
	req.WithTag(strPtr("pxe-enable"))
	if site != "" {
		req.WithSite(&site)
	}

	res, err := c.Dcim.DcimDevicesList(req, nil)
	if err != nil {
		log.Println("Error retreiving interface list", err)
		os.Exit(2)
	}

	hosts := make(map[int64]*DHCPHost, *res.Payload.Count)
	for _, dev := range res.Payload.Results {
		if dev.PrimaryIp4 == nil {
			log.Printf("Primary IPv4 unset on %s, skipping...", *dev.Name)
			continue
		}
		ipaddr := strings.SplitN(*dev.PrimaryIp4.Address, "/", 2)[0]

		ifreq := dcim.NewDcimInterfacesListParams()
		ifreq.WithDeviceID(strPtr(fmt.Sprintf("%d", dev.ID)))
		ifreq.WithMacAddressn(strPtr("null"))

		ifres, err := c.Dcim.DcimInterfacesList(ifreq, nil)
		if err != nil {
			log.Println("Error pulling interfaces for device", err)
			continue
		}

		if *ifres.Payload.Count == 0 {
			log.Printf("No interface available for PXE! (%s)", *dev.Name)
			continue
		}

		hwaddrs := make([]string, *ifres.Payload.Count)
		for i, int := range ifres.Payload.Results {
			if int.MacAddress == nil {
				continue
			}
			hwaddrs[i] = strings.ToLower(*int.MacAddress)
		}

		log.Println(dev.ID, *dev.Name, ipaddr, hwaddrs)
		hosts[dev.ID] = &DHCPHost{
			DeviceID: dev.ID,
			Name:     *dev.Name,
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
