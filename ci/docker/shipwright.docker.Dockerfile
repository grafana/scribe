ARG VERSION=latest
FROM shipwright:${VERSION}

RUN apk add docker
WORKDIR /var/shipwright
