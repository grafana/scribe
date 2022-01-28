ARG VERSION=latest
FROM shipwright:${VERSION}

RUN apk add --no-cache nodejs
WORKDIR /var/shipwright
