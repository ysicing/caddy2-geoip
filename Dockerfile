FROM golang:1-bookworm as builder

RUN go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

WORKDIR /go/src/github.com/ysicing/caddy2-geoip

COPY go.mod go.mod

COPY go.sum go.sum

RUN go mod download

COPY . .

RUN xcaddy build \
    --with github.com/ysicing/caddy2-geoip=../caddy2-geoip \
    --with github.com/caddy-dns/cloudflare  \
    --with github.com/caddy-dns/tencentcloud \
    --with github.com/caddy-dns/alidns \
    --with github.com/caddy-dns/dnspod \
    --with github.com/WeidiDeng/caddy-cloudflare-ip

FROM ysicing/debian

COPY --from=builder /go/src/github.com/ysicing/caddy2-geoip/caddy /usr/bin/caddy

COPY Caddyfile /etc/caddy/Caddyfile

ENV org.opencontainers.image.source = "https://github.com/ysicing/caddy2-geoip"

EXPOSE 2024

CMD caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
