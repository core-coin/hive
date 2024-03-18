package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/core-coin/go-core/v2/consensus"
	"github.com/core-coin/go-core/v2/consensus/cryptore"
	"github.com/core-coin/go-core/v2/core"
	"github.com/core-coin/go-core/v2/core/rawdb"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/core/vm"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
	"github.com/core-coin/go-core/v2/xcbdb"
	"golang.org/x/exp/slices"
)

// generatorConfig is the configuration of the chain generator.
type generatorConfig struct {
	// genesis options
	// forkInterval int    // number of blocks between forks
	// lastFork     string // last enabled fork
	// clique       bool   // create a clique chain

	// chain options
	txInterval  int // frequency of blocks containing transactions
	txCount     int // number of txs in block
	chainLength int // number of generated blocks

	// output options
	outputs   []string // enabled outputs
	outputDir string   // path where output files should be placed
}

func (cfg generatorConfig) withDefaults() (generatorConfig, error) {
	// if cfg.lastFork == "" {
	// 	cfg.lastFork = lastFork
	// }
	if cfg.txInterval == 0 {
		cfg.txInterval = 1
	}
	if cfg.outputs == nil {
		cfg.outputs = []string{"genesis", "chain", "txinfo"}
	}
	return cfg, nil
}

// generator is the central object in the chain generation process.
// It holds the configuration, state, and all instantiated transaction generators.
type generator struct {
	cfg      generatorConfig
	genesis  *core.Genesis
	td       *big.Int
	accounts []*crypto.PrivateKey
	rand     *rand.Rand

	// Modifier lists.
	virgins   []*modifierInstance
	mods      []*modifierInstance
	modOffset int

	// for write/export
	blockchain *core.BlockChain
}

type modifierInstance struct {
	name string
	blockModifier
}

func newGenerator(cfg generatorConfig) *generator {
	genesis := cfg.createGenesis()
	return &generator{
		cfg:      cfg,
		genesis:  genesis,
		rand:     rand.New(rand.NewSource(10)),
		td:       new(big.Int).Set(genesis.Difficulty),
		virgins:  cfg.createBlockModifiers(),
		accounts: slices.Clone(knownAccounts),
	}
}

func (cfg *generatorConfig) createBlockModifiers() (list []*modifierInstance) {
	for name, new := range modRegistry {
		list = append(list, &modifierInstance{
			name:          name,
			blockModifier: new(),
		})
	}
	slices.SortFunc(list, func(a, b *modifierInstance) int {
		return strings.Compare(a.name, b.name)
	})
	return list
}

// run produces a chain and writes it.
func (g *generator) run() error {
	// Init genesis block.
	db := rawdb.NewMemoryDatabase()
	genesis := g.genesis.MustCommit(db)
 
	// Create the blocks.
	chain, _ := core.GenerateChain(g.genesis.Config, genesis, cryptore.NewFaker(), db, g.cfg.chainLength, g.modifyBlock)

	// Import the chain. This runs all block validation rules.
	bc, err := g.importChain(cryptore.NewFaker(), chain, g.genesis.Config, genesis, db)
	if err != nil {
		return err
	}

	g.blockchain = bc
	return g.write()
}

func (g *generator) importChain(engine consensus.Engine, chain []*types.Block, config *params.ChainConfig, genesis *types.Block, db xcbdb.Database) (*core.BlockChain, error) {
	cacheconfig :=  &core.CacheConfig{
		TrieCleanLimit: 256,
		TrieDirtyLimit: 256,
		TrieTimeLimit:  5 * time.Minute,
		SnapshotLimit:  0, // Disable snapshot by default
		Preimages: true,
	}
	vmconfig := vm.Config{EnablePreimageRecording: true}
	blockchain, err := core.NewBlockChain(db, cacheconfig, config, engine, vmconfig, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("can't create blockchain: %v", err)
	}

	i, err := blockchain.InsertChain(chain)
	if err != nil {
		blockchain.Stop()
		return nil, fmt.Errorf("chain validation error (block %d): %v", chain[i].Number(), err)
	}
	return blockchain, nil
}

func (g *generator) modifyBlock(i int, gen *core.BlockGen) {
	fmt.Println("generating block", gen.Number())
	g.setDifficulty(i, gen)
	g.runModifiers(i, gen)
}

func (g *generator) setDifficulty(i int, gen *core.BlockGen) {
		prevBlock := gen.PrevBlock(-1)
		if prevBlock == nil {
			g.td = g.td.Add(g.td, big.NewInt(8192))
		} else {
			g.td = g.td.Add(g.td, prevBlock.Difficulty())
		}
}

// runModifiers executes the chain modifiers.
func (g *generator) runModifiers(i int, gen *core.BlockGen) {
	totalMods := len(g.mods) + len(g.virgins)
	if totalMods == 0 || g.cfg.txInterval == 0 || i%g.cfg.txInterval != 0 {
		return
	}

	ctx := &genBlockContext{index: i, block: gen, gen: g}

	// Modifier scheduling: we cycle through the available modifiers until enough have
	// executed successfully. It also stops when all of them return false from apply()
	// because this usually means there is no gas left.
	count := 0
	refused := 0 // count of consecutive times apply() returned false
	run := func(mod *modifierInstance) bool {
		ok := mod.apply(ctx)
		if ok {
			fmt.Println("    -", mod.name)
			count++
			refused = 0
		} else {
			refused++
		}
		return ok
	}

	// In order to avoid a pathological situation where a modifier never executes because
	// of unfortunate scheduling, we first try modifiers from g.virgins.
	for i := 0; i < len(g.virgins) && count < g.cfg.txCount; i++ {
		mod := g.virgins[i]
		if run(mod) {
			g.mods = append(g.mods, mod)
			g.virgins = append(g.virgins[:i], g.virgins[i+1:]...)
			i--
		}
	}
	// If there is any space left, fill it using g.mods.
	for len(g.mods) > 0 && count < g.cfg.txCount && refused < totalMods {
		index := g.modOffset % len(g.mods)
		run(g.mods[index])
		g.modOffset++
	}
}