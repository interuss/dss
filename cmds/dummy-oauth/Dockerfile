FROM golang:1.24-alpine AS build
RUN apk add git bash make
RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

RUN mkdir -p cmds/dummy-oauth
COPY cmds/dummy-oauth cmds/dummy-oauth

RUN go install ./...

FROM alpine:latest
COPY --from=build /go/bin/dummy-oauth /usr/bin
COPY build/test-certs /var/test-certs
ENTRYPOINT ["/usr/bin/dummy-oauth"]
