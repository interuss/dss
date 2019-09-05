FROM golang:1.12-alpine AS build
RUN apk add git bash make
RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

RUN mkdir -p cmds/dummy-oauth
COPY cmds/dummy-oauth cmds/dummy-oauth

RUN go install ./...

FROM alpine:latest
COPY --from=build /go/bin/dummy-oauth /usr/bin
ENTRYPOINT ["/usr/bin/dummy-oauth"]