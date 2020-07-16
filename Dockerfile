# This Dockerfile builds the InterUSS `dss` image which contains the binary
# executables for both the grpc-backend and the http-gateway.  To run a
# container for this image, the desired binary must be specified (either
# /usr/bin/grpc-backend or /usr/bin/http-gateway).

FROM golang:1.14.3-alpine AS build
RUN apk add git bash make
RUN mkdir /app
COPY go.mod go.sum /app/
WORKDIR /app

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

COPY . /app
RUN make interuss

FROM alpine:latest
RUN apk update && apk add ca-certificates
COPY --from=build /go/bin/http-gateway /usr/bin
COPY --from=build /go/bin/grpc-backend /usr/bin
