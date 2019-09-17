FROM golang:1.12-alpine AS build
RUN apk add git bash make
ADD . /app
WORKDIR /app
RUN go mod download
RUN make

FROM alpine:latest
RUN apk update && apk add ca-certificates
COPY --from=build /go/bin/http-gateway /usr/bin
ENTRYPOINT ["/usr/bin/http-gateway"]