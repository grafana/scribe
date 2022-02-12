FROM golang:1.17 as builder
WORKDIR /app
COPY . .

RUN go build \
    -ldflags \
    "-X main.Version=$(git describe --tags --dirty --always)" \
    -o bin/shipwright ./plumbing/cmd

FROM alpine:3
COPY --from=builder /app/bin/shipwright /bin/shipwright
RUN apk add --no-cache bash go

WORKDIR /var/shipwright
