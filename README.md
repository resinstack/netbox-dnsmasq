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
  * `NETBOX_PROTOCOL` - Connect the given protocol, defaults to https.
  * `NETBOX_HOST` - Address of the netbox host to connect to.
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
