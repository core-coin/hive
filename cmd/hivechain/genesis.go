package main

import (
	"math/big"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/params"
)

var initialBalance, _ = new(big.Int).SetString("1000000000000000000000000000000000000", 10)

// createChainConfig creates a chain configuration.
func (cfg *generatorConfig) createChainConfig() *params.ChainConfig {
	chaincfg := new(params.ChainConfig)

	networkid, _ := new(big.Int).SetString("3503995874084926", 10)
	chaincfg.NetworkID = networkid
	common.DefaultNetworkID = common.NetworkID(networkid.Int64())
	chaincfg.Cryptore = new(params.CryptoreConfig)

	return chaincfg
}

func (cfg *generatorConfig) genesisDifficulty() *big.Int {
	return new(big.Int).Set(params.MinimumDifficulty)
}

// createGenesis creates the genesis block and config.
func (cfg *generatorConfig) createGenesis() *core.Genesis {
	var g core.Genesis
	g.Config = cfg.createChainConfig()

	// Block attributes.
	g.Difficulty = cfg.genesisDifficulty()
	g.ExtraData = []byte("hivechain")
	g.Number = 0
	g.EnergyLimit = params.GenesisEnergyLimit * 8

	g.Coinbase = core.DefaultCoinbasePrivate
	if g.Config.NetworkID.Uint64() == 1 {
		g.Coinbase = core.DefaultCoinbaseMainnet
	} else if g.Config.NetworkID.Uint64() == 3{
		g.Coinbase = core.DefaultCoinbaseDevin
	}
	 
	// Initialize allocation.
	// Here we add balance to known accounts and initialize built-in contracts.
	g.Alloc = make(core.GenesisAlloc)
	for _, acc := range knownAccounts {
		g.Alloc[acc.Address()] = core.GenesisAccount{Balance: initialBalance}
	}
	addEmitContract(g.Alloc)

	return &g
}

const emitAddr = "7dcd17433742f4c0ca53122ab541d0ba67fc27df"

func addEmitContract(ga core.GenesisAlloc) {
	checksum := common.CalculateChecksum(common.Hex2Bytes(emitAddr), common.Hex2Bytes(common.DefaultNetworkID.String()))
	addr, err := common.HexToAddress(common.DefaultNetworkID.String()+ checksum+ emitAddr)
	if err != nil {
		panic(err)
	}
	ga[addr] = core.GenesisAccount{
		Balance: new(big.Int),
		Code:    emitCode,
	}
}