package main

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
)

type blockModifier interface {
	apply(*genBlockContext) bool
	txInfo() any
}

var modRegistry = make(map[string]func() blockModifier)

// register adds a block modifier.
func register(name string, new func() blockModifier) {
	modRegistry[name] = new
}

type genBlockContext struct {
	index   int
	block   *core.BlockGen
	gen     *generator
	txcount int
}

// Number returns the block number.
func (ctx *genBlockContext) Number() *big.Int {
	return ctx.block.Number()
}

// NumberU64 returns the block number.
func (ctx *genBlockContext) NumberU64() uint64 {
	return ctx.block.Number().Uint64()
}

// Timestamp returns the block timestamp.
func (ctx *genBlockContext) Timestamp() uint64 {
	block := ctx.block.PrevBlock(-1)
	return block.Time()
}

// HasEnergy reports whether the block still has more than the given amount of energy left.
func (ctx *genBlockContext) HasEnergy(energy uint64) bool {
	return ctx.block.Energy() > energy
}

// AddNewTx adds a transaction into the block.
func (ctx *genBlockContext) AddNewTx(sender *crypto.PrivateKey, tx *types.Transaction) *types.Transaction {
	signedTx, err := types.SignTx(tx, ctx.Signer(), sender)
	if err != nil {
		panic(err)
	}
	ctx.block.AddTx(signedTx)
	ctx.txcount++
	return signedTx
}

// TxSenderAccount chooses an account to send transactions from.
func (ctx *genBlockContext) TxSenderAccount() *crypto.PrivateKey {
	a := ctx.gen.accounts[0]
	return a
}

// TxCreateIntrinsicEnergy gives the 'intrinsic energy' of a contract creation transaction.
func (ctx *genBlockContext) TxCreateIntrinsicEnergy(data []byte) uint64 {
	ienergy, err := core.IntrinsicEnergy(data, true)
	if err != nil {
		panic(err)
	}
	return ienergy
}

// TxEnergyFeeCap returns the minimum energyprice that should be used for transactions.
func (ctx *genBlockContext) TxEnergyFeeCap() *big.Int {
	return big.NewInt(1)
}

// AccountNonce returns the current nonce of an address.
func (ctx *genBlockContext) AccountNonce(addr common.Address) uint64 {
	return ctx.block.TxNonce(addr)
}

// Signer returns a signer for the current block.
func (ctx *genBlockContext) Signer() types.Signer {
	return types.MakeSigner(ctx.ChainConfig().NetworkID)
}

// TxCount returns the number of transactions added so far.
func (ctx *genBlockContext) TxCount() int {
	return ctx.txcount
}

// ChainConfig returns the chain config.
func (ctx *genBlockContext) ChainConfig() *params.ChainConfig {
	return ctx.gen.genesis.Config
}

// ParentBlock returns the parent of the current block.
func (ctx *genBlockContext) ParentBlock() *types.Block {
	return ctx.block.PrevBlock(ctx.index - 1)
}

// TxRandomValue returns a random value that depends on the block number and current transaction index.
func (ctx *genBlockContext) TxRandomValue() uint64 {
	var txindex [8]byte
	binary.BigEndian.PutUint64(txindex[:], uint64(ctx.TxCount()))
	h := sha256.New()
	h.Write(ctx.Number().Bytes())
	h.Write(txindex[:])
	return binary.BigEndian.Uint64(h.Sum(nil))
}
