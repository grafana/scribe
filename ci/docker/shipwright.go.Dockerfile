ARG VERSION=latest
FROM ghcr.io/grafana/shipwright:${VERSION}

RUN apk add --no-cache go
WORKDIR /var/shipwright
