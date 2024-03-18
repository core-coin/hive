#!/bin/bash

#
# This is a little test-script, that can be used for a some trial runs of clients.
#
# This script runs all production-ready tests, but does so with very restrictive options,
# and should thus complete in tens of minutes
#

HIVEHOME="./"

# Store results in temp
RESULTS="/tmp/TestResults"

FLAGS="--loglevel 4"
FLAGS="$FLAGS --results-root $RESULTS "
FLAGS="$FLAGS --sim.parallelism 1 --client.checktimelimit=20s"

echo "Running the quick'n'dirty version of the Hive tests, for local development"
echo "To get the hive viewer up, you can do"
echo ""
echo "  cd $HIVEHOME/hiveviewer && ln -s /tmp/TestResults/ Results && python -m SimpleHTTPServer"
echo ""
echo "And then visit http://localhost:8000/ with your browser. "
echo "Log-files and stuff is availalbe in $RESULTS."
echo ""
echo ""


function run {
  echo "$HIVEHOME> $1"
  (cd $HIVEHOME && $1)
}

function testgraphql {
  echo "$(date) Starting graphql simulation [$1]"
  run "./hive --sim core-coin/graphql --client $1 $FLAGS"
}

function testsync {
  echo "$(date) Starting hive sync simulation [$1]"
  run "./hive --sim core-coin/sync --client=$1 $FLAGS"
}

function testdevp2p {
  echo "$(date) Starting p2p simulation [$1]"
  run "./hive --sim devp2p --client $1 $FLAGS"
}

mkdir $RESULTS

# main tests (devp2p, sync)
testdevp2p go-core_latest
testgraphql go-core_latest
testsync go-core_latest

# smoke tests
./hive --sim smoke/genesis --client go-core_latest --loglevel 4 --results-root /tmp/TestResults  --sim.parallelism 1 --client.checktimelimit=60s
./hive --sim smoke/mining --client go-core_latest --loglevel 4 --results-root /tmp/TestResults  --sim.parallelism 1 --client.checktimelimit=60s
./hive --sim smoke/network --client go-core_latest --loglevel 4 --results-root /tmp/TestResults  --sim.parallelism 1 --client.checktimelimit=60s

# # rpc tests
./hive --sim core-coin/rpc --client go-core_latest --loglevel 4 --results-root /tmp/TestResults  --sim.parallelism 1 --client.checktimelimit=120s
