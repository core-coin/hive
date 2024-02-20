package main

import (
	"math/big"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/params"
)

func init() {
	register("tx-transfer", func() blockModifier {
		return &modValueTransfer{
			energyLimit: params.TxEnergy,
		}
	})
}

type modValueTransfer struct {
	energyLimit uint64

	txs []valueTransferInfo
}

type valueTransferInfo struct {
	TxHash common.Hash    `json:"txhash"`
	Sender common.Address `json:"sender"`
	Block  hexutil.Uint64 `json:"block"`
	Index  int            `json:"indexInBlock"`
}

func (m *modValueTransfer) apply(ctx *genBlockContext) bool {
	if !ctx.HasEnergy(m.energyLimit) {
		return false
	}

	sender := ctx.TxSenderAccount()
	recipient := pickRecipient(ctx)

	txdata := types.NewTransaction(ctx.AccountNonce(sender.Address()), recipient, big.NewInt(1), m.energyLimit, ctx.TxEnergyFeeCap(), nil)

	txindex := ctx.TxCount()
	tx := ctx.AddNewTx(sender, txdata)
	m.txs = append(m.txs, valueTransferInfo{
		Block:  hexutil.Uint64(ctx.NumberU64()),
		Sender: sender.Address(),
		TxHash: tx.Hash(),
		Index:  txindex,
	})
	return true
}

func (m *modValueTransfer) txInfo() any {
	return m.txs
}

func pickRecipient(ctx *genBlockContext) common.Address {
	i := ctx.TxRandomValue() % uint64(len(ctx.gen.accounts))
	return ctx.gen.accounts[i].Address()
}
