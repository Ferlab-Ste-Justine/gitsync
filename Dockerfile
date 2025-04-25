FROM golang:1.23-bullseye AS builder

ENV CGO_ENABLED=0

WORKDIR /opt

COPY . /opt/

RUN go mod tidy

RUN go build .

FROM scratch

COPY --from=builder /opt/gitsync /gitsync

ENTRYPOINT ["/gitsync"]