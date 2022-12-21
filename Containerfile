FROM docker.io/golang:1.19-alpine as build
WORKDIR /netbox-dnsmasq
COPY . .
RUN go mod vendor && go build -o /netbox-dhcp-hosts .

FROM docker.io/alpine:3.17 as base
WORKDIR /
COPY --from=build /netbox-dhcp-hosts /usr/local/bin/netbox-dhcp-hosts
COPY runit/dnsmasq /etc/service/dnsmasq
COPY runit/netbox-dnsmasq-dhcp /etc/service/netbox-dnsmasq-dhcp
RUN apk update && \
    apk add tini runit dnsmasq && \
    rm -rf /var/cache/apk && \
    mkdir -p /var/lib/tftp && \
    wget -O /var/lib/tftp/ipxe.efi https://boot.ipxe.org/ipxe.efi
ENTRYPOINT ["/sbin/tini", "/sbin/runsvdir", "/etc/service"]

FROM docker.io/golang:1.19-alpine as shoelaces_build
WORKDIR /shoelaces
RUN apk add git && \
    git clone -b v1.2.0 https://github.com/thousandeyes/shoelaces.git . && \
    go mod vendor && go build -o ./shoelaces .

FROM base as shoelaces
WORKDIR /
COPY --from=shoelaces_build /shoelaces/shoelaces /usr/local/bin/shoelaces
COPY --from=shoelaces_build /shoelaces/web /usr/share/shoelaces/web
COPY runit/shoelaces /etc/service/shoelaces
RUN mkdir -p /var/lib/shoelaces && printf "---\nnetworkMaps:\n" > /var/lib/shoelaces/mappings.yaml
ENV BIND_ADDR=0.0.0.0:8081
ENTRYPOINT ["/sbin/tini", "/sbin/runsvdir", "/etc/service"]
