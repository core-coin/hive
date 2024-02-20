package main

import (
	"github.com/core-coin/go-core/v2/core/types"
)

// Here we create transactions that create spam contracts. These exist simply to fill up
// the state. We need a decent amount of state in the sync tests, for example.

func init() {
	register("randomlogs", func() blockModifier {
		return &modCreateSpam{
			code: genlogsCode,
			energy:  20000,
		}
	})
	register("randomcode", func() blockModifier {
		return &modCreateSpam{
			code: gencodeCode,
			energy:  60000,
		}
	})
	register("randomstorage", func() blockModifier {
		return &modCreateSpam{
			code: genstorageCode,
			energy:  80000,
		}
	})
}

type modCreateSpam struct {
	code []byte
	energy  uint64
}

func (m *modCreateSpam) apply(ctx *genBlockContext) bool {
	energy := ctx.TxCreateIntrinsicEnergy(m.code) + m.energy
	if !ctx.HasEnergy(energy) {
		return false
	}

	sender := ctx.TxSenderAccount()
	//todo:error2215 add some `to` address and amount 
	tx := types.NewTransaction(ctx.AccountNonce(sender.Address()), sender.Address(), nil, energy, ctx.TxEnergyFeeCap(), m.code)
	ctx.AddNewTx(sender, tx)
	return true
}

func (m *modCreateSpam) txInfo() any {
	return nil
}
