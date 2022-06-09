ARG VERSION=latest
FROM grafana/shipwright:${VERSION}

RUN apk add --no-cache git openssh
WORKDIR /var/scribe
