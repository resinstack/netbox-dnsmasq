FROM docker.io/golang:1.19-alpine as build
WORKDIR /netbox-dnsmasq
COPY . .
RUN go mod vendor && go build -o /netbox-dhcp-hosts .

FROM docker.io/alpine:3.17
COPY --from=build /netbox-dhcp-hosts /usr/local/bin/netbox-dhcp-hosts
COPY runit /etc/service
RUN apk update && \
    apk add tini runit dnsmasq && \
    rm -rf /var/cache/apk && \
    mkdir -p /var/lib/tftp && \
    wget -O /var/lib/tftp/ipxe.efi https://boot.ipxe.org/ipxe.efi

ENTRYPOINT ["/sbin/tini", "/sbin/runsvdir", "/etc/service"]
