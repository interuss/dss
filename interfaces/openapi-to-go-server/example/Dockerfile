FROM golang:1.22-alpine

WORKDIR /go/src/server_demo
COPY . .

ENV CGO_ENABLED=0
RUN go get -d -v ./...
RUN go install -v -a

CMD ["example"]
