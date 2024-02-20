# Build the simulator.
FROM golang:1-alpine AS builder
RUN apk --no-cache add gcc musl-dev linux-headers g++
ADD . /mining
WORKDIR /mining
RUN go build .

# Build the runner container.
FROM alpine:latest
ADD . /
COPY --from=builder /mining/mining /
ENTRYPOINT ["./mining"]
