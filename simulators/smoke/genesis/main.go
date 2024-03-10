package main

import (
	"github.com/core-coin/hive/hivesim"
)

func main() {
	suite := hivesim.Suite{
		Name:        "genesis",
		Description: "This test suite checks client initialization with genesis blocks.",
	}
	suite.Add(hivesim.ClientTestSpec{
		Role:        "xcb1",
		Name:        "empty genesis",
		Description: "This imports an empty genesis block.",
		Files: map[string]string{
			"/genesis.json": "genesis-empty.json",
		},
		Parameters: map[string]string{
			"HIVE_NETWORK_ID":            "7",
		},
		Run: genesisTest{"0x709b9d43c547920818628cc46d0908c2158fc3d99ce5b74b397a4344fcf62e7c"}.test,
	})
	suite.Add(hivesim.ClientTestSpec{
		Role:        "xcb1",
		Name:        "non-empty",
		Description: "This imports a non-empty genesis block.",
		Files: map[string]string{
			"/genesis.json": "genesis-nonempty.json",
		},
		Parameters: map[string]string{
			"HIVE_NETWORK_ID":            "7",
		},
		Run: genesisTest{"0x483845f05cd7d26a38a818e3cf04007deb1ce508f95eaf1810c8528f48fddfc0"}.test,
	})
	suite.Add(hivesim.ClientTestSpec{
		Name:        "precomp-storage",
		Description: "This imports a genesis where a precompile has code/nonce/storage.",
		Files: map[string]string{
			"/genesis.json": "genesis-precomp-storage.json",
		},
		Parameters: map[string]string{
			"HIVE_NETWORK_ID":            "7",
		},
		Run: genesisTest{"0xfc2f5fb65cdd077cf0c8172b45cac2a304f570a71de2046333abe5832080954a"}.test,
	})
	suite.Add(hivesim.ClientTestSpec{
		Name:        "precomp-empty",
		Description: "This imports a genesis where a precompile is an empty account.",
		Files: map[string]string{
			"/genesis.json": "genesis-precomp-empty.json",
		},
		Parameters: map[string]string{
			"HIVE_NETWORK_ID":            "7",
		},
		Run: genesisTest{"0xd2d343dbfb3e243681046bd708c60770c55688cd9fff9e047e10f9e48e1d310d"}.test,
	})
	suite.Add(hivesim.ClientTestSpec{
		Role:        "xcb1",
		Name:        "precomp-zero-balance",
		Description: "This imports a genesis where a precompile has code/nonce/storage, but balance is zero.",
		Files: map[string]string{
			"/genesis.json": "genesis-precomp-zero-balance.json",
		},
		Parameters: map[string]string{
			"HIVE_NETWORK_ID":            "7",
		},
		Run: genesisTest{"0x139047cc6de920174727c048897fa6ba144187473496392e472c819f7b5d5de2"}.test,
	})

	hivesim.MustRunSuite(hivesim.New(), suite)
}

type block struct {
	Hash string `json:"hash"`
}

type genesisTest struct {
	wantHash string
}

func (g genesisTest) test(t *hivesim.T, c *hivesim.Client) {
	var b block
	if err := c.RPC().Call(&b, "xcb_getBlockByNumber", "0x0", false); err != nil {
		t.Fatal("xcb_getBlockByNumber call failed:", err)
	}
	t.Log("genesis hash", b.Hash)
	if b.Hash != g.wantHash {
		t.Fatal("wrong genesis hash, want", g.wantHash)
	}
}
