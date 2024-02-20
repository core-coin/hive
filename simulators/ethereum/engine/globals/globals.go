package globals

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/hive/hivesim"
)

type TestAccount struct {
	key     *crypto.PrivateKey
	index   uint64
}

func (a *TestAccount) GetKey() *crypto.PrivateKey {
	return a.key
}

func (a *TestAccount) GetAddress() common.Address {
	return a.key.Address()
}

func (a *TestAccount) GetIndex() uint64 {
	return a.index
}

var (

	// Test chain parameters
	NetworkID          = big.NewInt(7)
	EnergyPrice         = big.NewInt(30 * params.Nucle)
	EnergyTipPrice      = big.NewInt(1 * params.Nucle)
	BlobEnergyPrice     = big.NewInt(1 * params.Nucle)
	GenesisTimestamp = uint64(0x1234)

	// RPC Timeout for every call
	RPCTimeout = 10 * time.Second

	// Engine, Eth ports
	XcbPortHTTP    = 8545
	EnginePortHTTP = 8551

	// JWT Authentication Related
	DefaultJwtTokenSecretBytes = []byte("secretsecretsecretsecretsecretse") // secretsecretsecretsecretsecretse
	MaxTimeDriftSeconds        = int64(60)

	// Accounts used for testing
	TestAccountCount = uint64(1000)
	TestAccounts     []*TestAccount

	// Global test case timeout
	DefaultTestCaseTimeout = time.Second * 60

	// Confirmation blocks
	PoWConfirmationBlocks = uint64(15)
	PoSConfirmationBlocks = uint64(1)

	MinerPKHex   = "44cc42018b039b182dc1f8a05f4696995e39ba0e2af6d14f8dc68ee2fdf77195a85cbd6f8cf94905015b039300f44d8d639b287cb15abe4db0"
	MinerAddrHex = "cb65e49851f010cd7d81b5b4969f3b0e8325c415359d"

	DefaultClientEnv = hivesim.Params{
		"HIVE_NETWORK_ID":          NetworkID.String(),
		"HIVE_PRIVATEKEY": MinerPKHex,
		"HIVE_MINER":             MinerAddrHex,
	}
)

func init() {
	// Fill the test accounts with deterministic addresses
	TestAccounts = make([]*TestAccount, TestAccountCount)
	for i := uint64(0); i < TestAccountCount; i++ {
		k, err := crypto.GenerateKey(rand.Reader)
		if err != nil {
			panic(err)
		}
		TestAccounts[i] = &TestAccount{key: k, index: i}
	}
}
