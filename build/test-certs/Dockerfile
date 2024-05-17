FROM golang:1.22-alpine

WORKDIR /go/src/app

ADD . .
RUN go mod init
RUN go build  -o gen-cert

CMD ["./gen-cert"]
