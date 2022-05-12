FROM golang:1.17-alpine
ENV PATH="${GOPATH}/bin:${PATH}"
RUN apk add bash git protoc protobuf-dev
RUN mkdir /build
WORKDIR /build
COPY go.mod .
RUN go mod download
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.14.3
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.14.3
RUN go install github.com/NYTimes/openapi2proto/cmd/openapi2proto
RUN go install github.com/golang/protobuf/protoc-gen-go@v1.4.1
