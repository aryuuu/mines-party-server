FROM golang:1.19-alpine

WORKDIR /go/src/app
# copy src
COPY . .
# install deps
RUN go get -d -v ./...
# compile binary
RUN go build -o mines-party-server ./cmd/server
# run binary
CMD ["./mines-party-server"]


