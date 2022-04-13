# This Dockerfile builds the InterUSS `dss` image which contains the binary
# executables for the core-service, http-gateway and the db-manager. It also
# contains a light weight tool that provides debugging capability. To run a
# container for this image, the desired binary must be specified (either
# /usr/bin/core-service, /usr/bin/http-gateway or /usr/bin/db-manager).

FROM golang:1.17-alpine AS build
RUN apk add build-base
RUN apk add git bash make
RUN mkdir /app
COPY go.mod go.sum /app/
# Intend to run delve download outside the go module directory to prevent it
# from being added as a dependency
RUN go install github.com/go-delve/delve/cmd/dlv@v1.8.2
WORKDIR /app

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

COPY .git /app/.git
COPY cmds /app/cmds
RUN mkdir -p cmds/db-manager

COPY pkg /app/pkg
COPY cmds/db-manager cmds/db-manager

RUN go install ./...

COPY scripts /app/scripts
COPY Makefile /app
RUN make interuss


FROM alpine:latest
RUN apk update && apk add ca-certificates
COPY --from=build /go/bin/http-gateway /usr/bin
COPY --from=build /go/bin/core-service /usr/bin
COPY --from=build /go/bin/db-manager /usr/bin
COPY --from=build /go/bin/dlv /usr/bin
HEALTHCHECK CMD cat service.ready || exit 1
