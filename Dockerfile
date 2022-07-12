FROM golang:1.18-bullseye as builder
WORKDIR /go/src/app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ .

RUN go get -d -v ./...
RUN go build .

CMD ./juscaba-to-telegram
