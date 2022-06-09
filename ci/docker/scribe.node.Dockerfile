ARG VERSION=latest
FROM grafana/shipwright:${VERSION}

RUN apk add --no-cache nodejs
WORKDIR /var/scribe
