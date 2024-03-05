package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/hexutil"
	"github.com/core-coin/go-core/v2/core"
	"golang.org/x/exp/maps"
)

var outputFunctions = map[string]func(*generator) error{
	"genesis":        (*generator).writeGenesis,
	"forkenv":        (*generator).writeForkEnv,
	"chain":          (*generator).writeChain,
	"powchain":       (*generator).writePoWChain,
	"headstate":      (*generator).writeState,
	"headblock":      (*generator).writeHeadBlock,
	"accounts":       (*generator).writeAccounts,
	"txinfo":         (*generator).writeTxInfo,
}

func outputFunctionNames() []string {
	names := maps.Keys(outputFunctions)
	sort.Strings(names)
	return names
}

// write creates the generator output files.
func (g *generator) write() error {
	var wf []func(*generator) error
	for _, name := range g.cfg.outputs {
		fmt.Println("writing", name)
		f := outputFunctions[name]
		if f == nil {
			return fmt.Errorf("unknown output %q", name)
		}
		wf = append(wf, f)
	}
	for _, f := range wf {
		if err := f(g); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) openOutputFile(file string) (*os.File, error) {
	path := filepath.Join(g.cfg.outputDir, file)
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

func (g *generator) writeJSON(name string, obj any) error {
	jsonData, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	out, err := g.openOutputFile(name)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.Write(jsonData)
	return err
}

// writeGenesis writes the genesis.json file.
func (g *generator) writeGenesis() error {
	return g.writeJSON("genesis.json", g.genesis)
}

// writeAccounts writes the account keys file.
func (g *generator) writeAccounts() error {
	type accountObj struct {
		Key string `json:"key"`
	}
	m := make(map[common.Address]*accountObj, len(g.accounts))
	for _, a := range g.accounts {
		m[a.Address()] = &accountObj{
			Key: hexutil.Encode(a.PrivateKey()),
		}
	}
	return g.writeJSON("accounts.json", &m)
}

// writeState writes the chain state dump.
func (g *generator) writeState() error {
	headstate, err := g.blockchain.State()
	if err != nil {
		return err
	}
	//todo:error2215 maybe we need to exclude some data
	dump := headstate.RawDump(false, false, false)
	return g.writeJSON("headstate.json", &dump)
}

// writeHeadBlock writes information about the head block.
func (g *generator) writeHeadBlock() error {
	return g.writeJSON("headblock.json", g.blockchain.CurrentHeader())
}

// writeChain writes all RLP blocks to a file.
func (g *generator) writeChain() error {
	path := filepath.Join(g.cfg.outputDir, "chain.rlp")
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	lastBlock := g.blockchain.CurrentBlock().NumberU64()
	return exportN(g.blockchain, out, 1, lastBlock)
}

// writePoWChain writes pre-merge RLP blocks to a file.
func (g *generator) writePoWChain() error {
	path := filepath.Join(g.cfg.outputDir, "powchain.rlp")
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	lastBlock := g.blockchain.CurrentBlock().NumberU64()

	return exportN(g.blockchain, out, 1, lastBlock)
}

func exportN(bc *core.BlockChain, w io.Writer, first uint64, last uint64) error {
	for nr := first; nr <= last; nr++ {
		block := bc.GetBlockByNumber(nr)
		if block == nil {
			return fmt.Errorf("export failed on #%d: not found", nr)
		}
		if err := block.EncodeRLP(w); err != nil {
			return err
		}
	}
	return nil
}

// writeTxInfo writes information about the transactions that were added into the chain.
func (g *generator) writeTxInfo() error {
	m := make(map[string]any, len(g.mods))
	for _, inst := range g.mods {
		m[inst.name] = inst.txInfo()
	}
	return g.writeJSON("txinfo.json", &m)
}
