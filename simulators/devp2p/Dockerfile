# This simulator runs devp2p protocol tests.

# Build devp2p tool.
FROM golang:1-alpine as gocore-builder
RUN apk add --update git gcc musl-dev linux-headers g++
RUN git clone --depth 1 https://github.com/core-coin/go-core.git /go-core
WORKDIR /go-core
RUN go build -v ./cmd/devp2p

# Build the simulator executable.
FROM golang:1-alpine as sim-builder
RUN apk add --update git gcc musl-dev linux-headers g++
WORKDIR /source
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o devp2p-simulator

# Build the simulation run container.
FROM alpine:latest
ADD . /source
WORKDIR /source
COPY --from=gocore-builder /go-core/devp2p ./devp2p
COPY --from=gocore-builder /go-core/cmd/devp2p/internal/xcbtest/testdata /testchain
COPY --from=sim-builder /source/devp2p-simulator .
ENTRYPOINT ["./devp2p-simulator"]
