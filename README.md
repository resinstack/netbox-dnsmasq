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
  * `DNSMASQ_HOSTSFILE` - A file to write out the dhcp hosts
    configuration to.  This must match wherever you configure your
    dnsmasq config file to search for dhcp hosts.
  * `SHOELACES_MAPFILE` - A file to write out shoelaces mappings too.
    Must be named `mappings.yaml` and at the path expected by
    shoelaces.  Only relevant in images that contain shoelaces.
  * `SHOELACES_TAG_PREFIX` - A prefix that if found on a tag will be
    used to form the mapping for shoelaces to supply the correct boot
    files to the machine without human interaction.

The search through netbox by default pulls hosts that have the
$NETBOX_TAG tag set.  Hosts are then filtered to ensure they have a
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

## Using with Shoelaces

Shoelaces is a powerful ipxe mapping tool that can mux different boot
images to different machines by hand or by address.  Each build of the
containers from this repo comes in a non-shoelaces flavor and a
shoelaces flavor.  Using shoelaces will provide a web interface on
port 8081 that allows you to manually select ipxe scripts.  Shoelaces
expects to find your template scripts in `/var/lib/shoelaces/ipxe` and
can serve static assets for you from `/var/lib/shoelaces/static`.

Shoelaces is also capable of directly mapping a machine to its boot
scripts based on tags in netbox.  To use this behavior configure the
value of `SHOELACES_TAG_PREFIX` to the prefix used by the tags you
have set in Netbox.  For example, if you wanted machines that posess
the tags `pxe-zerotouch-deb11` to boot with the ipxe script
`deb11.ipxe`, you would configure the tag to be `pxe-zerotouch-`. Only
the first tag found will be mapped, so its a good idea to carefully
plan the tag structure in use for these tags to be unique.
