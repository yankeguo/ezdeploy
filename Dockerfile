FROM golang:1.19 AS builder
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -o /ezops ./cmd/ezops

FROM alpine:3.16

RUN apk add --no-cache ca-certificates tzdata curl && \
    curl -sSL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/v1.24.7/bin/linux/amd64/kubectl" && \
    chmod +x /usr/local/bin/kubectl && \
    curl -sSL -o helm.tar.gz "https://get.helm.sh/helm-v3.10.1-linux-amd64.tar.gz" && \
    tar xf helm.tar.gz && \
    mv -f linux-amd64/helm /usr/local/bin/helm && \
    rm -rf helm.tar.gz linux-amd64

COPY --from=builder /ezops /ezops

WORKDIR /data