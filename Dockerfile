FROM docker.io/golang:1.23-alpine AS build
WORKDIR /netbox-dnsmasq
COPY . .
RUN go mod vendor && go build -o /netbox-dhcp-hosts .

FROM docker.io/alpine:3.21 AS ipxe-build
ARG IPXE_REV=eeec6442
WORKDIR /ipxe
RUN apk update && apk add git build-base perl && \
    git clone https://github.com/ipxe/ipxe.git . && \
    git checkout ${IPXE_REV} && cd src && \
    make bin-x86_64-efi/ipxe.efi && \
    make bin-x86_64-efi/snponly.efi

FROM docker.io/alpine:3.21 AS base
WORKDIR /
COPY --from=build /netbox-dhcp-hosts /usr/local/bin/netbox-dhcp-hosts
COPY runit/dnsmasq /etc/service/dnsmasq
COPY runit/netbox-dnsmasq-dhcp /etc/service/netbox-dnsmasq-dhcp
ENV DNSMASQ_HOSTSFILE=/run/dhcp-hosts.next
ENV REFRESH_INTERVAL=600
RUN apk update && \
    apk add tini runit dnsmasq && \
    rm -rf /var/cache/apk && \
    mkdir -p /var/lib/tftp
COPY --from=ipxe-build /ipxe/src/bin-x86_64-efi/ipxe.efi /var/lib/tftp/ipxe.efi
COPY --from=ipxe-build /ipxe/src/bin-x86_64-efi/snponly.efi /var/lib/tftp/snponly.efi
ENTRYPOINT ["/sbin/tini", "/sbin/runsvdir", "/etc/service"]

FROM docker.io/golang:1.23-alpine AS shoelaces_build
WORKDIR /shoelaces
ARG SHOELACES_PRS=
ARG SHOELACES_ORIGIN=https://github.com/thousandeyes/shoelaces
RUN apk add git curl && \
    git clone -b v1.3.2 https://github.com/thousandeyes/shoelaces.git . && \
        git config --global user.email "build@localhost" && \
        git config --global user.name "Shoelaces Build" && \
    for p in ${SHOELACES_PRS} ; do curl -L ${SHOELACES_ORIGIN}/pull/$p.patch | git am - ; done && \
    go mod vendor && go build -o ./shoelaces .

FROM base AS shoelaces
WORKDIR /
ENV SHOELACES_MAPFILE=/var/lib/shoelaces/mappings.yaml.next
COPY --from=shoelaces_build /shoelaces/shoelaces /usr/local/bin/shoelaces
COPY --from=shoelaces_build /shoelaces/web /usr/share/shoelaces/web
COPY runit/shoelaces /etc/service/shoelaces
RUN mkdir -p /var/lib/shoelaces && printf "---\nnetworkMaps:\n" > /var/lib/shoelaces/mappings.yaml
ENV BIND_ADDR=0.0.0.0:8081
ENTRYPOINT ["/sbin/tini", "/sbin/runsvdir", "/etc/service"]
