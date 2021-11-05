# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

ADD . /go/src/main
WORKDIR /go/src/main
RUN go get main
RUN go install
ENTRYPOINT ["/go/bin/main"]

EXPOSE 8081
