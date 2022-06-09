ARG VERSION=latest
FROM grafana/shipwright:${VERSION}

RUN apk add docker git
WORKDIR /var/scribe
