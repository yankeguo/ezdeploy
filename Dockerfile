FROM golang:1.22 AS builder
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -o /ezdeploy ./cmd/ezdeploy

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl && \
    curl -sSL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/v1.28.8/bin/linux/amd64/kubectl" && \
    chmod +x /usr/local/bin/kubectl && \
    curl -sSL -o helm.tar.gz "https://get.helm.sh/helm-v3.14.3-linux-amd64.tar.gz" && \
    mkdir -p helm && \
    tar -xf helm.tar.gz -C helm --strip-components 1 && \
    mv helm/helm /usr/local/bin/helm && \
    rm -rf helm.tar.gz helm

COPY --from=builder /ezdeploy /usr/local/bin/ezdeploy

ENTRYPOINT ["ezdeploy"]

WORKDIR /data
