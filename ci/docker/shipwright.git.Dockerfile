ARG VERSION=latest
FROM ghcr.io/grafana/shipwright:${VERSION}

RUN apk add --no-cache git openssh
WORKDIR /var/shipwright
