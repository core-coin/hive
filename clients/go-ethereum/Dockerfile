ARG baseimage=core-coin/go-core
ARG tag=latest

FROM $baseimage:$tag as builder

FROM alpine:latest
RUN apk add --update bash curl jq
COPY --from=builder /usr/local/bin/gocore /usr/local/bin/gocore

# Generate the version.txt file.
RUN /usr/local/bin/gocore console --exec 'console.log(admin.nodeInfo.name)' --maxpeers=0 --nodiscover --dev 2>/dev/null | head -1 > /version.txt

# Inject the startup script.
ADD gocore.sh /gocore.sh
ADD mapper.jq /mapper.jq
RUN chmod +x /gocore.sh

# Inject the enode id retriever script.
RUN mkdir /hive-bin
ADD enode.sh /hive-bin/enode.sh
RUN chmod +x /hive-bin/enode.sh

# Add a default genesis file.
ADD genesis.json /genesis.json

# Export the usual networking ports to allow outside access to the node
EXPOSE 8545 8546 8547 8551 30300 30300/udp

ENTRYPOINT ["/gocore.sh"]
