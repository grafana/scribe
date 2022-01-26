FROM golang:1.17 as builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:3
COPY --from=builder /app/bin/shipwright /bin/shipwright
