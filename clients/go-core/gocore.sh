#!/bin/bash

# Startup script to initialize and boot a go-core instance.
#
# This script assumes the following files:
#  - `gocore` binary is located in the filesystem root
#  - `genesis.json` file is located in the filesystem root (mandatory)
#  - `chain.rlp` file is located in the filesystem root (optional)
#  - `blocks` folder is located in the filesystem root (optional)
#  - `keys` folder is located in the filesystem root (optional)
#
# This script assumes the following environment variables:
#
#  - HIVE_BOOTNODE                enode URL of the remote bootstrap node
#  - HIVE_NETWORK_ID              network ID number to use for the xcb protocol
#  - HIVE_NODETYPE                sync and pruning selector (archive, full, light)
#  - HIVE_PRIVATEKEY              private key to use for signing and mining
#
# Other:
#
#  - HIVE_MINER                   enable mining. value is coinbase address.
#  - HIVE_MINER_EXTRA             extra-data field to set for newly minted blocks
#  - HIVE_LOGLEVEL                client loglevel (0-5)
#  - HIVE_GRAPHQL_ENABLED         enables graphql on port 8545
#  - HIVE_LES_SERVER              set to '1' to enable LES server

# Immediately abort the script on any error encountered
set -e

gocore=/usr/local/bin/gocore
FLAGS=""

if [ "$HIVE_LOGLEVEL" != "" ]; then
    FLAGS="$FLAGS --verbosity=$HIVE_LOGLEVEL"
fi

# It doesn't make sense to dial out, use only a pre-set bootnode.
FLAGS="$FLAGS --bootnodes=$HIVE_BOOTNODE"

# If a specific network ID is requested, use that
if [ "$HIVE_NETWORK_ID" != "" ]; then
    FLAGS="$FLAGS --networkid $HIVE_NETWORK_ID"
else
    # Unless otherwise specified by hive, we try to avoid mainnet networkid. If gocore detects mainnet network id,
    # then it tries to bump memory quite a lot
    FLAGS="$FLAGS --networkid 1337"
    HIVE_NETWORK_ID=1337
    echo $HIVE_NETWORK_ID
fi

# Handle any client mode or operation requests
if [ "$HIVE_NODETYPE" == "archive" ]; then
    FLAGS="$FLAGS --syncmode full --gcmode archive"
fi
if [ "$HIVE_NODETYPE" == "full" ]; then
    FLAGS="$FLAGS --syncmode full"
fi
if [ "$HIVE_NODETYPE" == "light" ]; then
    FLAGS="$FLAGS --syncmode light"
fi
if [ "$HIVE_NODETYPE" == "" ]; then
    FLAGS="$FLAGS --syncmode full"
fi

# Configure the chain.
mv /genesis.json /genesis-input.json
jq -f /mapper.jq /genesis-input.json > /genesis.json

# Dump genesis. 
if [ "$HIVE_LOGLEVEL" -lt 4 ]; then
    echo "Supplied genesis state (trimmed, use --sim.loglevel 4 or 5 for full output):"
    jq 'del(.alloc[] | select(.balance == "0x123450000000000000000"))' /genesis.json
else
    echo "Supplied genesis state:"
    cat /genesis.json
fi

# Initialize the local testchain with the genesis state
echo "Initializing database with genesis state..."
$gocore $FLAGS init /genesis.json

# Don't immediately abort, some imports are meant to fail
set +e

# Load the test chain if present
echo "Loading initial blockchain..."
if [ -f /chain.rlp ]; then
    $gocore $FLAGS import /chain.rlp
else
    echo "Warning: chain.rlp not found."
fi

# Load the test halfchain if present
echo "Loading halfchain..."
if [ -f /halfchain.rlp ]; then
    $gocore $FLAGS import /halfchain.rlp
else
    echo "Warning: halfchain.rlp not found."
fi

# Load the remainder of the test chain
echo "Loading remaining individual blocks..."
if [ -d /blocks ]; then
    (cd /blocks && $gocore $FLAGS --gcmode=archive --verbosity=$HIVE_LOGLEVEL --nocompaction import `ls | sort -n`)
else
    echo "Warning: blocks folder not found."
fi

set -e

# Import signing key.
if [ "$HIVE_PRIVATEKEY" != "" ]; then
    # Create password file.
    echo "Importing key..."
    echo "foobar" > /gocore-password-file.txt
    $gocore account import --networkid $HIVE_NETWORK_ID  --password /gocore-password-file.txt <(echo "$HIVE_PRIVATEKEY")

    # Ensure password file is used when running gocore in mining mode.
    if [ "$HIVE_MINER" != "" ]; then
        FLAGS="$FLAGS --password /gocore-password-file.txt --unlock $HIVE_MINER --allow-insecure-unlock"
    fi
fi

# Configure any mining operation
if [ "$HIVE_MINER" != "" ] && [ "$HIVE_NODETYPE" != "light" ]; then
    FLAGS="$FLAGS --mine --miner.corebase $HIVE_MINER --miner.threads=99"
fi
if [ "$HIVE_MINER_EXTRA" != "" ]; then
    FLAGS="$FLAGS --miner.extradata $HIVE_MINER_EXTRA"
fi

# # Configure LES.
if [ "$HIVE_LES_SERVER" == "1" ]; then
  FLAGS="$FLAGS --light.serve 80 --bttp "
fi

# Configure RPC.
FLAGS="$FLAGS --http --http.addr=0.0.0.0 --http.port=8545 --http.api=admin,debug,xcb,miner,net,personal,txpool,web3"
FLAGS="$FLAGS --ws --ws.addr=0.0.0.0 --ws.origins \"*\" --ws.api=admin,debug,xcb,miner,net,personal,txpool,web3"

if [ "$HIVE_TERMINAL_TOTAL_DIFFICULTY" != "" ]; then
    echo "0x7365637265747365637265747365637265747365637265747365637265747365" > /jwtsecret
    FLAGS="$FLAGS --authrpc.addr=0.0.0.0 --authrpc.port=8551 --authrpc.jwtsecret /jwtsecret"
fi

# Configure GraphQL.
if [ "$HIVE_GRAPHQL_ENABLED" != "" ]; then
    FLAGS="$FLAGS --graphql"
fi

# Run the go-core implementation with the requested flags.
FLAGS="$FLAGS --nat=none"
echo "Running go-core with flags $FLAGS"
$gocore $FLAGS
