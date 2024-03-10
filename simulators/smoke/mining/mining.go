package main

import (
	"time"

	"github.com/core-coin/hive/hivesim"
)

func main() {
	suite := hivesim.Suite{
		Name:        "mining",
		Description: "This test suite tests mining support.",
	}
	suite.Add(hivesim.ClientTestSpec{
		Role:        "xcb1",
		Name:        "mine one block",
		Description: "Waits for a single block to get mined.",
		Files: map[string]string{
			"/genesis.json": "genesis.json",
		},
		Parameters: hivesim.Params{
			"HIVE_PRIVATEKEY": "44cc42018b039b182dc1f8a05f4696995e39ba0e2af6d14f8dc68ee2fdf77195a85cbd6f8cf94905015b039300f44d8d639b287cb15abe4db0",
			"HIVE_MINER":             "ce56e49851f010cd7d81b5b4969f3b0e8325c415359d",
			"HIVE_NETWORK_ID":            "1334",
		},
		Run: miningTest,
	})
	hivesim.MustRunSuite(hivesim.New(), suite)
}

type block struct {
	Number     string `json:"number"`
	Hash       string `json:"hash"`
	ParentHash string `json:"parentHash"`
}

func miningTest(t *hivesim.T, c *hivesim.Client) {
	start := time.Now()
	timeout := 20 * time.Second

	for {
		var b block
		if err := c.RPC().Call(&b, "xcb_getBlockByNumber", "latest", false); err != nil {
			t.Fatal("xcb_getBlockByNumber call failed:", err)
		}
		switch b.Number {
		case "0x0":
			// still at genesis block, keep waiting.
			if time.Since(start) > timeout {
				t.Fatal("no block produced within", timeout)
			}
			time.Sleep(300 * time.Millisecond)
		case "0x1":
			t.Log("block mined:", b.Hash)
			return
		default:
			t.Fatal("wrong latest block: number", b.Number, ", hash", b.Hash, ", parent", b.ParentHash)
		}
	}
}
