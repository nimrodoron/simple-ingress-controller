FROM golang AS builder
MAINTAINER "Nimrod Oron <nimrod.oron@sap.com>"

RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential && \
    apt-get clean && \
    mkdir -p "$GOPATH/src/github.com/nimrodoron/simple-ingress-controller"

ADD . "$GOPATH/src/github.com/nimrodoron/simple-ingress-controller"

RUN cd "$GOPATH/src/github.com/nimrodoron/simple-ingress-controller" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a --installsuffix cgo --ldflags="-s" -o /simple-ingress-controller

ENTRYPOINT ["/simple-ingress-controller"]