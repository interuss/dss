# This Dockerfile builds the InterUSS `dss` image which contains the binary
# executables for both the core-service and the http-gateway. It also
# contains a light weight tool that provides debugging capability. To run a
# container for this image, the desired binary must be specified (either
# /usr/bin/core-service or /usr/bin/http-gateway).

FROM golang:1.14.3-alpine AS build
RUN apk add git bash make
RUN mkdir /app
COPY go.mod go.sum /app/
# Intend to run delve download outside the go module directory to prevent it
# from being added as a dependency
RUN go get github.com/go-delve/delve/cmd/dlv
WORKDIR /app

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

COPY .git /app/.git
COPY cmds /app/cmds
COPY pkg /app/pkg
COPY scripts /app/scripts
COPY Makefile /app
RUN make interuss

FROM alpine:latest
RUN apk update && apk add ca-certificates
COPY --from=build /go/bin/http-gateway /usr/bin
COPY --from=build /go/bin/core-service /usr/bin
COPY --from=build /go/bin/dlv /usr/bin
