ARG VERSION=latest
FROM grafana/shipwright:${VERSION}

RUN apk add --no-cache go git
WORKDIR /var/scribe
