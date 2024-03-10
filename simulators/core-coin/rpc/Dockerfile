# This simulation runs JSON-RPC API tests.
FROM golang:1-alpine as builder
RUN apk add --update gcc musl-dev linux-headers g++

# Build the simulator executable.
ADD . /source
WORKDIR /source
RUN go build -v .

# Build the simulator run container.
FROM alpine:latest
RUN apk add --update curl git gcc musl-dev linux-headers g++ make
ADD . /source
WORKDIR /source
COPY --from=builder /source/rpc .
ENTRYPOINT ["./rpc"]
