package helper

import (
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/trie"
)

var (
	EmptyWithdrawalsRootHash = &types.EmptyRootHash
)

func ComputeWithdrawalsRoot(ws types.Withdrawals) common.Hash {
	// Using RLP root but might change to ssz
	return types.DeriveSha(
		ws,
		trie.NewStackTrie(nil),
	)
}
