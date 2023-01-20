# syntax=docker/dockerfile:1.3-labs

FROM golang:1.19 AS base

RUN apt-get update
RUN apt-get install -y ca-certificates

ENV GO111MODULE=on
ENV GOPROXY=direct
ENV GOSUMDB=off
ENV CGO_ENABLED=0

WORKDIR /root

COPY netrc ./.netrc
RUN chmod 600 /root/.netrc

COPY dockerconfig/gitconfig ./.gitconfig:

RUN mkdir -p ./.ssh/
COPY dockerconfig/ssh_config ./.ssh/config

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

FROM base AS builder

ARG VERSION
ARG GOARCH
ARG GOOS

RUN go mod download

RUN rm -rf /root/.netrc

COPY main.go ./
COPY ./bfc-api ./bfc-api
COPY ./slack-sdk ./slack-sdk

RUN GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-X main.VERSION=${VERSION}" -o /bfc-bin-collection-notifier .

FROM debian:stable-slim AS final

COPY --from=builder /bfc-bin-collection-notifier /bfc-bin-collection-notifier
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/bfc-bin-collection-notifier"]
