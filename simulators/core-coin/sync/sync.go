package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/xcbclient"
	"github.com/core-coin/hive/hivesim"
)

var (
	// the number of seconds before a sync is considered stalled or failed
	syncTimeout = 60 * time.Second
	sourceFiles = map[string]string{
		"genesis.json": "./chain/genesis.json",
		"chain.rlp":    "./chain/chain.rlp",
	}
	sinkFiles = map[string]string{
		"genesis.json": "./chain/genesis.json",
	}
)

func main() {
	// Load fork environment.
	var params hivesim.Params
	err := common.LoadJSON("chain/forkenv.json", &params)
	if err != nil {
		panic(err)
	}
	networkID, err  := strconv.Atoi(params["HIVE_NETWORK_ID"])
	if err != nil {
		panic("Bad network ID in forkenv.json")
	}
	common.DefaultNetworkID = common.NetworkID(networkID)
	var suite = hivesim.Suite{
		Name: "sync",
		Description: `This suite of tests verifies that clients can sync from each other in different modes.
For each client, we test if it can serve as a sync source for all other clients (including itself).`,
	}
	suite.Add(hivesim.ClientTestSpec{
		Role:        "xcb1",
		Name:        "CLIENT as sync source",
		Description: "This loads the test chain into the client and verifies whether it was imported correctly.",
		Parameters:  params,
		Files:       sourceFiles,
		Run: func(t *hivesim.T, c *hivesim.Client) {
			runSourceTest(t, c, params)
		},
	})
	hivesim.MustRunSuite(hivesim.New(), suite)
}

func runSourceTest(t *hivesim.T, c *hivesim.Client, params hivesim.Params) {
	// Check whether the source has imported its chain.rlp correctly.
	source := &node{c}
	if err := source.checkHead(); err != nil {
		t.Fatal(err)
	}

	// Configure sink to connect to the source node.
	enode, err := source.EnodeURL()
	if err != nil {
		t.Fatal("can't get node peer-to-peer endpoint:", enode)
	}
	sinkParams := params.Set("HIVE_BOOTNODE", enode)

	// Sync all sink nodes against the source.
	t.RunAllClients(hivesim.ClientTestSpec{
		Role:        "xcb1",
		Name:        fmt.Sprintf("sync %s -> CLIENT", source.Type),
		Description: fmt.Sprintf("This test attempts to sync the chain from a %s node.", source.Type),
		Parameters:  sinkParams,
		Files:       sinkFiles,
		Run:         runSyncTest,
	})
}

func runSyncTest(t *hivesim.T, c *hivesim.Client) {
	node := &node{c}
	err := node.checkSync(t)
	if err != nil {
		t.Fatal("sync failed:", err)
	}
}

type node struct {
	*hivesim.Client
}

// checkSync waits for the node to reach the head of the chain.
func (n *node) checkSync(t *hivesim.T) error {
	var expectedHead types.Header
	err := common.LoadJSON("chain/headblock.json", &expectedHead)
	if err != nil {
		return fmt.Errorf("can't load expected header: %v", err)
	}
	wantHash := expectedHead.Hash()

	// if err := n.triggerSync(t); err != nil {
	// 	return err
	// }

	var (
		timeout = time.After(syncTimeout)
		current = uint64(0)
	)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout (%v elapsed, current head is %d)", syncTimeout, current)
		default:
			block, err := n.head()
			if err != nil {
				t.Logf("error getting block from %s (%s): %v", n.Type, n.Container, err)
				return err
			}
			blockNumber := block.Number.Uint64()
			if blockNumber != current {
				t.Logf("%s has new head %d", n.Type, blockNumber)
			}
			if current == expectedHead.Number.Uint64() {
				if block.Hash() != wantHash {
					return fmt.Errorf("wrong head hash %x, want %x", block.Hash(), wantHash)
				}
				return nil // success
			}
			// check in a little while....
			current = blockNumber
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

// checkHead checks whether the remote chain head matches the given values.
func (n *node) checkHead() error {
	var expected types.Header
	err := common.LoadJSON("chain/headblock.json", &expected)
	if err != nil {
		return fmt.Errorf("can't load expected header: %v", err)
	}

	head, err := n.head()
	if err != nil {
		return fmt.Errorf("can't query chain head: %v", err)
	}
	if head.Hash() != expected.Hash() {
		return fmt.Errorf("wrong chain head %d (%s), want %d (%s)", head.Number, head.Hash().TerminalString(), expected.Number, expected.Hash().TerminalString())
	}
	return nil
}

// head returns the node's chain head.
func (n *node) head() (*types.Header, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return xcbclient.NewClient(n.RPC()).HeaderByNumber(ctx, nil)
}