ARG VERSION=latest
FROM shipwright:${VERSION}

RUN apk add --no-cache go
WORKDIR /var/shipwright
