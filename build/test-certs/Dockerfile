FROM golang:1.24-alpine

WORKDIR /go/src/app

ADD . .
RUN go mod init
RUN go build  -o gen-cert

CMD ["./gen-cert"]
