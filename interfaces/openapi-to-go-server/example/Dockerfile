FROM golang:1.17-alpine

WORKDIR /go/src/server_demo
COPY . .

RUN go install -d -v ./...
RUN go install -v -a

CMD ["example"]
