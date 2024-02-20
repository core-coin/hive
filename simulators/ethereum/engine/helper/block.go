package helper

import (
	"fmt"

	"github.com/core-coin/go-core/v2/core/types"
)

func GenerateInvalidPayloadBlock(baseBlock *types.Block, uncle *types.Block) (*types.Block, error) {
		if uncle == nil {
			return nil, fmt.Errorf("No ommer provided")
		}
		uncles := []*types.Header{
			uncle.Header(),
		}
		newHeader := types.CopyHeader(baseBlock.Header())
		newHeader.UncleHash = types.CalcUncleHash(uncles)

		modifiedBlock := types.NewBlockWithHeader(newHeader).WithBody(baseBlock.Transactions(), uncles)
		fmt.Printf("DEBUG: hash=%s, ommerLen=%d, ommersHash=%v\n", modifiedBlock.Hash(), len(modifiedBlock.Uncles()), modifiedBlock.UncleHash())
		return modifiedBlock, nil
}
