FROM golang:1.22-alpine AS build
RUN apk add git bash make
RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

RUN mkdir -p cmds/token-signer
COPY cmds/token-signer .

RUN go install ./...

FROM alpine:latest
COPY --from=build /go/bin/token-signer /usr/bin
COPY build/test-certs /var/test-certs
COPY cmds/token-signer/.env .
ENTRYPOINT ["/usr/bin/token-signer"]