FROM golang:1.18 as builder
WORKDIR /app
COPY . .

RUN go build \
    -ldflags \
    "-X main.Version=$(git describe --tags --dirty --always)" \
    -o bin/scribe ./plumbing/cmd

FROM alpine:edge
COPY --from=builder /app/bin/scribe /bin/scribe
RUN apk update && apk add --no-cache bash go

WORKDIR /var/scribe
