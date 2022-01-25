FROM golang:1.17 as builder
WORKDIR /app
COPY . .
RUN make build

FROM debian:11-slim
COPY --from=builder /app/shipwright /bin/shipwright
RUN apt update -yq
RUN apt install -y curl build-essential git ssh
# Docker recommended / best-practice for apt repositories
RUN rm -rf /var/lib/apt/lists/*
