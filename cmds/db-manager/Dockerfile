FROM golang:1.12-alpine AS build
RUN apk add git bash make
RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

RUN mkdir -p cmds/db-manager
COPY pkg pkg
COPY cmds/db-manager cmds/db-manager

RUN go install ./...

FROM alpine:latest
COPY --from=build /go/bin/db-manager /usr/bin
ENTRYPOINT ["/usr/bin/db-manager"]
