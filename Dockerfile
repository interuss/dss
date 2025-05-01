# This Dockerfile builds the InterUSS `dss` image which contains the binary
# executables for the core-service and the db-manager. It also
# contains a light weight tool that provides debugging capability. To run a
# container for this image, the desired binary must be specified (either
# /usr/bin/core-service or /usr/bin/db-manager).

FROM golang:1.24-alpine AS build
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

# In order to reliably compute the version of the build, all files must be present.
# This is required to detect a dirty workspace using `scripts/git/version.sh`.
COPY . /app
RUN make interuss


FROM alpine:latest
RUN apk update && apk add ca-certificates
COPY --from=build /go/bin/core-service /usr/bin
COPY --from=build /go/bin/db-manager /usr/bin
COPY --from=build /go/bin/dlv /usr/bin
COPY build/jwt-public-certs /jwt-public-certs
COPY build/test-certs /test-certs
COPY build/db_schemas /db-schemas
HEALTHCHECK CMD cat service.ready || exit 1
