package main

import (
	"encoding/binary"
	"math/big"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/log"
)

func init() {
	register("tx-emit", func() blockModifier {
		return &modInvokeEmit{
			energyLimit: 100000,
		}
	})
}

// modInvokeEmit creates transactions that invoke the 'emit' contract.
type modInvokeEmit struct {
	energyLimit uint64

	txs []invokeEmitTxInfo
}

type invokeEmitTxInfo struct {
	TxHash    common.Hash    `json:"txhash"`
	Sender    common.Address `json:"sender"`
	Block     hexutil.Uint64 `json:"block"`
	Index     int            `json:"indexInBlock"`
	LogTopic0 common.Hash    `json:"logtopic0"`
	LogTopic1 common.Hash    `json:"logtopic1"`
}

func (m *modInvokeEmit) apply(ctx *genBlockContext) bool {
	if !ctx.HasEnergy(m.energyLimit) {
		return false
	}

	sender := ctx.TxSenderAccount()
	recipient, err := common.HexToAddress(emitAddr)
	if err != nil {
		log.Error("failed to parse emit address", "err", err)
		return false
	}
	calldata := m.genCallData(ctx)
	datahash := crypto.SHA3Hash(calldata)

	txdata := types.NewTransaction(ctx.AccountNonce(sender.Address()), recipient, big.NewInt(2), m.energyLimit, ctx.TxEnergyFeeCap(), calldata)
	txindex := ctx.TxCount()
	tx := ctx.AddNewTx(sender, txdata)
	m.txs = append(m.txs, invokeEmitTxInfo{
		Block:     hexutil.Uint64(ctx.NumberU64()),
		Sender:    sender.Address(),
		TxHash:    tx.Hash(),
		Index:     txindex,
		LogTopic0: common.HexToHash("0x00000000000000000000000000000000000000000000000000000000656d6974"),
		LogTopic1: datahash,
	})
	return true
}

func (m *modInvokeEmit) txInfo() any {
	return m.txs
}

func (m *modInvokeEmit) genCallData(ctx *genBlockContext) []byte {
	d := make([]byte, 8)
	binary.BigEndian.PutUint64(d, ctx.TxRandomValue())
	return append(d, "emit"...)
}
