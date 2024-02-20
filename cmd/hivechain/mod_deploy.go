package main

import (
	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
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
	gas := ctx.TxCreateIntrinsicEnergy(code)
	gas += uint64(len(m.code)) * params.CreateDataEnergy
	gas += 15000 // extra gas for constructor execution
	if !ctx.HasEnergy(gas) {
		return false
	}

	sender := ctx.TxSenderAccount()
	nonce := ctx.AccountNonce(sender.Address())
	//todo:error2215 add some `to` address and amount 
	tx := types.NewTransaction(nonce, sender.Address(), nil, gas, ctx.TxEnergyFeeCap(), code)
	ctx.AddNewTx(sender, tx)
	m.info = &deployTxInfo{
		Contract: crypto.CreateAddress(sender.Address(), nonce),
		Block:    hexutil.Uint64(ctx.block.Number().Uint64()),
	}
	return true
}

func (m *modDeploy) txInfo() any {
	return m.info
}
