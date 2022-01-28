ARG VERSION=latest
FROM ghcr.io/grafana/shipwright:${VERSION}

RUN apk add --no-cache nodejs
WORKDIR /var/shipwright
