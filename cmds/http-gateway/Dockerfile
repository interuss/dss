FROM golang:1.12-alpine AS build
RUN apk add git bash
ADD . /app
WORKDIR /app
RUN go mod download
RUN make

FROM alpine:latest
COPY --from=build /go/bin/grpc-backend /usr/bin
ENTRYPOINT ["/usr/bin/http-gateway"]