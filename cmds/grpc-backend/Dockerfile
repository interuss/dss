FROM golang:1.12-alpine AS build
RUN apk add git bash make
RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

RUN mkdir pkg
COPY pkg pkg

RUN mkdir -p cmds/grpc-backend
COPY cmds/grpc-backend cmds/grpc-backend

RUN go install ./...

FROM alpine:latest
RUN apk update && apk add ca-certificates
COPY --from=build /go/bin/grpc-backend /usr/bin
ENTRYPOINT ["/usr/bin/grpc-backend"]