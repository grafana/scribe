ARG VERSION=latest
FROM ghcr.io/grafana/shipwright:${VERSION}

RUN apk add docker git
WORKDIR /var/shipwright
