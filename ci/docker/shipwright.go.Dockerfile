ARG VERSION=latest
FROM grafana/shipwright:${VERSION}

RUN apk add --no-cache go
WORKDIR /var/shipwright
