[Overview] | [Hive Commands] | [Simulators] | [Clients]

## Hive Clients

This page explains how client containers work in Hive.

Clients are docker images which can be instantiated by a simulation. A client definition
consists of a Dockerfile and associated resources. Client definitions live in
subdirectories of `clients/` in the hive repository.

See the [go-core client definition][gocore-docker] for an example of a client
Dockerfile.

When hive runs a simulation, it first builds all client docker images using their
Dockerfile, i.e. it basically runs `docker build .` in the client directory. Since most
client definitions wrap an existing Core Blockchain client, and building the client from source
may take a long time, it is usually best to base the hive client wrapper on a pre-built
docker image from Docker Hub.

The client Dockerfile should support an optional argument named `branch`, which specifies
the requested client version. This argument can be set by users by appending it to the
client name like:

    ./hive --sim my-simulation --client go-core_v2.1.8,go_core_v2.1.9

Other build arguments can also be set using a YAML file, see the [hive command
documentation][hive-client-yaml] for more information.

### Alternative Dockerfiles

There can be other Dockerfiles besides the main one. Typically, a client should also
provide a `Dockerfile.git` that builds the client from source code. Alternative
Dockerfiles can be selected through hive's `-client-file` YAML configuration.

### hive.yaml

Hive reads additional metadata from the `hive.yaml` file in the client directory (next to
the Dockerfile). Currently, the only purpose of this file is specifying the client's role
list:

    roles:
      - "xcb1"
      - "xcb1_light_client"

The role list is available to simulators and can be used to differentiate between clients
based on features. Declaring a client role also signals that the client supports certain
role-specific environment variables and files. If `hive.yaml` is missing or doesn't declare
roles, the `xcb1` role is assumed.

### /version.txt

Client Dockerfiles are expected to generate a `/version.txt` file during build. Hive reads
this file after building the container and attaches version information to the output of
all test suites in which the client is launched.

### /hive-bin

Executables placed into the `/hive-bin` directory of the client container can be invoked
through the simulation API.

## Client Lifecycle

When the simulation requests a client instance, hive creates a docker container from the
client image. The simulator can customize the container by passing environment variables
with prefix `HIVE_`. It may also upload files into the container before it starts. Once
the container is created, hive simply runs the entry point defined in the `Dockerfile`.

For all client containers, hive waits for TCP port 8545 to open before considering the
client ready for use by the simulator. This port is configurable through the
`HIVE_CHECK_LIVE_PORT` variable, and the check can be disabled by setting it to `0`. If
the client container does not open this port within a certain timeout, hive assumes the
client has failed to start.

Environment variables and files interpreted by the entry point define a 'protocol' between
the simulator and client. While hive itself does not require support for any specific
variables or files, simulators usually expect client containers to be configurable in
certain ways. In order to run tests against multiple Core Blockchain clients, for example, the
simulator needs to be able to configure all clients for a specific blockchain and make
them join the peer-to-peer network used for testing.

## Xcb1 Client Requirements

This section describes the requirements for the `xcb1` client role.

Xcb1 clients must provide JSON-RPC over HTTP on TCP port 8545. They may also support
JSON-RPC over WebSocket on port 8546, but this is not strictly required.

### Files

The simulator customizes client startup by placing these files into the xcb1 client
container:

- `/genesis.json` contains Core Blockchain genesis state in the JSON format used by Gocore. This
  file is mandatory.
- `/chain.rlp` contains RLP-encoded blocks to import before startup.
- `/blocks/` directory containing `.rlp` files.

On startup, the entry point script must first load the genesis block and state into the
client implementation from `/genesis.json`. To do this, the script needs to translate from
Gocore genesis format into a format appropriate for the specific client implementation. The
translation is usually done using a jq script. See the [go-core genesis
translator][gocore-genesis-jq], for example.

After the genesis state, the client should import the blocks from `/chain.rlp` if it is
present, and finally import the individual blocks from `/blocks` in file name order. The
reason for requiring two different block sources is that specifying a single chain is more
optimal, but tests requiring forking chains cannot create a single chain. The client
should start even if the blocks are invalid, i.e. after the import, the client's 'best
block' should be the last valid, imported block.

### Scripts

Some tests require peer-to-peer node information of the client instance. All xcb1 client
containers must contain a `/hive-bin/enode.sh` script. This script should output the enode
URL of the running instance.

### Environment

Clients must support the following environment variables. The client's entry point script
may map these to command line flags or use them to generate a config file, for example.

| Variable                   | Value         |                                                |
|----------------------------|---------------|------------------------------------------------|
| `HIVE_LOGLEVEL`            | 0 - 5         | configures log level of client                 |
| `HIVE_NODETYPE`            | archive, full | sets sync algorithm                            |
| `HIVE_BOOTNODE`            | enode URL     | makes client connect to another node           |
| `HIVE_GRAPHQL_ENABLED`     | 0 - 1         | if set, GraphQL is enabled on port 8545        |
| `HIVE_MINER`               | address       | if set, mining is enabled. value is coinbase   |
| `HIVE_MINER_EXTRA`         | hex           | extradata for mined blocks                     |
| `HIVE_PRIVATEKEY`          | hex           | private key for signing                        |
| `HIVE_NETWORK_ID`          | decimal       | p2p network ID                                 |


## LES client/server roles

Xcb1 clients containing an implementation of [LES] may additionally support roles
`xcb1_les_client` and `xcb1_les_server`.

For the client role, the following additional variables should be supported:

| Variable                   | Value         |                                                |
|----------------------------|---------------|------------------------------------------------|
| `HIVE_NODETYPE`            | "light"       | enables LES client mode                        |

For the server role, the following additional variables should be supported:

| Variable                   | Value         |                                                |
|----------------------------|---------------|------------------------------------------------|
| `HIVE_LES_SERVER`          | 0 - 1         | if set to 1, LES server should be enabled      |


[LES]: https://github.com/ethereum/devp2p/blob/master/caps/les.md
[gocore-docker]: ../clients/go-core/Dockerfile
[hive-client-yaml]: ./commandline.md#client-build-parameters
[gocore-genesis-jq]: ../clients/go-core/mapper.jq
[Overview]: ./overview.md
[Hive Commands]: ./commandline.md
[Simulators]: ./simulators.md
[Clients]: ./clients.md
