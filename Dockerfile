FROM golang:1.18.7-alpine3.15 as builder

WORKDIR /app

# copy modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# cache modules
RUN go mod download

# copy source code
COPY main.go main.go
COPY pkg/ pkg/
# build
RUN CGO_ENABLED=0 go build \
    -a -o mycsi main.go

FROM alpine:3.13

WORKDIR /app

USER root
COPY --from=builder --chown=nobody:nobody /app/mycsi .

ENTRYPOINT ["./mycsi"]