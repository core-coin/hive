package main

import (
	"fmt"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/trie"
)

func init() {
	register("uncles", func() blockModifier {
		return &modUncles{
			info: make(map[uint64]unclesInfo),
		}
	})
}

type modUncles struct {
	info    map[uint64]unclesInfo
	counter int
}

type unclesInfo struct {
	Hashes []common.Hash `json:"hashes"`
}

func (m *modUncles) apply(ctx *genBlockContext) bool {
	info := m.info[ctx.NumberU64()]
	if len(info.Hashes) >= 2 {
		return false // block has enough uncles already
	}

	parent := ctx.ParentBlock()
	time := parent.Time() + 1
	uncle := &types.Header{
		Number:     parent.Number(),
		ParentHash: parent.ParentHash(),
		Time:       time,
		Extra:      []byte(fmt.Sprintf("hivechain uncle %d", m.counter)),
	}
	// Initialize the remaining remaining header fields by converting to a full block.
	ub := types.NewBlock(uncle, nil, nil, nil, trie.NewStackTrie(nil))
	uncle = ub.Header()

	// Add the uncle to the generated block.
	// Note that AddUncle computes the difficulty and gas limit for us.
	ctx.block.AddUncle(uncle)

	info.Hashes = append(info.Hashes, uncle.Hash())
	m.info[ctx.NumberU64()] = info
	m.counter++
	return true
}

func (m *modUncles) txInfo() any {
	return m.info
}
