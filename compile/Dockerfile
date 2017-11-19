FROM golang:alpine
WORKDIR /go/src/app

RUN apk add --no-cache git

ENTRYPOINT go-wrapper download && go-wrapper install && go build -o controller main.go