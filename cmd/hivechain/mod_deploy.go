package main

import (
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/params"
)

func init() {
	register("deploy-callme", func() blockModifier {
		return &modDeploy{code: callmeCode}
	})
	register("deploy-callenv", func() blockModifier {
		return &modDeploy{code: callenvCode}
	})
	register("deploy-callrevert", func() blockModifier {
		return &modDeploy{code: callrevertCode}
	})
}

type modDeploy struct {
	code []byte
	info *deployTxInfo
}

type deployTxInfo struct {
	Contract common.Address `json:"contract"`
	Block    hexutil.Uint64 `json:"block"`
}

func (m *modDeploy) apply(ctx *genBlockContext) bool {
	if m.info != nil {
		return false // already deployed
	}

	var code []byte
	code = append(code, deployerCode...)
	code = append(code, m.code...)
	gas := ctx.TxCreateIntrinsicGas(code)
	gas += uint64(len(m.code)) * params.CreateDataGas
	gas += 15000 // extra gas for constructor execution
	if !ctx.HasGas(gas) {
		return false
	}

	sender := ctx.TxSenderAccount()
	nonce := ctx.AccountNonce(sender.addr)
	ctx.AddNewTx(sender, &types.LegacyTx{
		Nonce:    nonce,
		Gas:      gas,
		GasPrice: ctx.TxGasFeeCap(),
		Data:     code,
	})
	m.info = &deployTxInfo{
		Contract: crypto.CreateAddress(sender.addr, nonce),
		Block:    hexutil.Uint64(ctx.block.Number().Uint64()),
	}
	return true
}

func (m *modDeploy) txInfo() any {
	return m.info
}
