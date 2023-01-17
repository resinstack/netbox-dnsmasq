# netbox-dnsmasq

This tool can be used to improve your machine provisioning system by
mirroring the IP address assignments from netbox to dnsmasq.  This
enables you to directly boot a machine to its correct address rather
than booting to a temporary environment with an ephemeral address and
having to re-assign the address later.

The following configuration variables are supported which may be set
in the environment:

  * `VERBOSE` - Enable logging on standard error.
  * `NETBOX_SITE` - Filter to only the given site, rather than all
    global data.
  * `NETBOX_TOKEN` - Token with suitable permissions to read netbox
    data.
  * `NETBOX_TAG` - Only service hosts that have a specific tag.
  * `NETBOX_URL` - URL to the netbox server.
  * `DNSMASQ_TEMPLATE` - A go template expression for the dhcp-hosts
    file.  Defaults to a suitable configuration for IPv4.  The default
    template is `{{JoinStrings .HWAddr ","}},{{.Addr}}`.

The search through netbox by default pulls hosts that have the
`pxe-enable` tag set.  Hosts are then filtered to ensure they hae a
primary IPv4 address and at least one interface that has a MAC address
set.  Hosts that have the correct tag but do not have at least one MAC
address associated with a non-management-only interface and a primary
address are skipped.  You can see which hosts are skipped by enabling
verbose logging.

The primary address is made available to all MAC addresses that are
not associated with a management-only interface due to the way netbox
models the network.  Typically the network is modeled from the
perspective of a booted and initialized machine, which may make use of
bonds, interface teaming, or other advanced interface topologies.  At
install time, it is more likely that only one interface will be
brought up, but it is still prefered that the machine get its final
address.


## Configuring dnsmasq

You will need to supply a suitable dnsmasq config file that points to
the auto-generated dhcp-hosts file.  Here is an example config that is
known to work:

```
keep-in-foreground
log-facility=-
enable-tftp
port=0
listen-address=10.0.0.10
tftp-root=/var/lib/tftp
dhcp-option=option:T1,300
dhcp-option=option:T2,525
dhcp-option=option:router,10.0.0.1
dhcp-option=option:dns-server,10.0.0.1
dhcp-option=option:netmask,255.255.0.0
dhcp-option=option:domain-name,resinstack.io
dhcp-userclass=set:ipxe,iPXE
dhcp-boot=tag:!ipxe,ipxe.efi
dhcp-boot=tag:ipxe,http://10.0.0.10:8081/poll/1/${netX/mac:hexhyp}
dhcp-range=10.0.0.10,static
dhcp-hostsfile=/run/dhcp-hosts
domain=resinstack.io
```

In this example the server hosting the PXE services is located at
`10.0.0.10` and the network has a default route and DNS available from
`10.0.0.1`.  The `,static` token in the dhcp-range block instructs
dnsmasq not to provide services to hosts that do not have a
pre-existing reservation, which can be useful if your network does not
use DHCP except for machine installation.
