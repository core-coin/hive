package chaingenerators

import (
	"github.com/core-coin/go-core/core/types"
	el "github.com/ethereum/hive/simulators/eth2/common/config/execution"
)

type ChainGenerator interface {
	Generate(*el.ExecutionGenesis) ([]*types.Block, error)
}
