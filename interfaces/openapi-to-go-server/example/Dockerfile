FROM golang:1.14

WORKDIR /go/src/server_demo
COPY . .

RUN go get -d -v ./...
RUN go install -v -a

CMD ["example"]
