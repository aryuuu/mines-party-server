FROM golang:1.19-alpine AS builder

WORKDIR /go/src/app
# copy src
COPY . .
# install deps
RUN go get -d -v ./...
# compile binary
RUN go build -o mines-party-server ./cmd/server

FROM alpine:latest AS base

COPY --from=builder /go/src/app/mines-party-server /mines-party-server
CMD ["/mines-party-server"]
