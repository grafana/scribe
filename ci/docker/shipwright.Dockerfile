FROM golang:1.17 as builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:3
COPY --from=builder /app/shipwright /bin/shipwright
