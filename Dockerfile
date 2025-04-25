FROM golang:1.23-bullseye AS builder

ENV CGO_ENABLED=0

WORKDIR /opt

COPY . /opt/

RUN go mod tidy

RUN go build .

FROM alpine:latest AS certs
RUN apk add --no-cache ca-certificates

FROM scratch

COPY --from=builder /opt/gitsync /gitsync
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/gitsync"]