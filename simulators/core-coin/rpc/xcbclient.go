package main

import (
	"bytes"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/core-coin/go-core/v2"
	"github.com/core-coin/go-core/v2/accounts/abi"
	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/crypto"
	"github.com/core-coin/go-core/v2/params"
)

var (
	contractCode = `
	pragma solidity >0.4.6;

contract Test {
    event E0();
    event E1(uint);
    event E2(uint indexed);
    event E3(address);
    event E4(address indexed);
    event E5(uint, address) anonymous;

    uint public ui;
    mapping(address => uint) map;

    constructor(uint ui_) {
        ui = ui_;
        map[msg.sender] = ui_;
    }

    function events(uint ui_, address addr_) public {
        emit E0();
        emit E1(ui_);
        emit E2(ui_);
        emit E3(addr_);
        emit E4(addr_);
        emit E5(ui_, addr_);
    }

    function constFunc(uint a, uint b, uint c) public view returns(uint, uint, uint) {
            return (a, b, c);
    }

    function getFromMap(address addr) public view returns(uint) {
        return map[addr];
    }

    function addToMap(address addr, uint value) public {
        map[addr] = value;
    }
}
	`
	// test contract deploy code, will deploy the contract with 1234 as argument
	deployCode = common.Hex2Bytes("608060405234801561001057600080fd5b5060405161067c38038061067c8339818101604052810190610032919061009c565b8060008190555080600160003375ffffffffffffffffffffffffffffffffffffffffffff1675ffffffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550506100e6565b600081519050610096816100cf565b92915050565b6000602082840312156100ae57600080fd5b60006100bc84828501610087565b91505092915050565b6000819050919050565b6100d8816100c5565b81146100e357600080fd5b50565b610587806100f56000396000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c80630beac9b51461005c57806345c8ce261461008c578063498059b9146100a8578063e38fe5c9146100da578063f62fb0e9146100f6575b600080fd5b6100766004803603810190610071919061031c565b610114565b6040516100839190610445565b60405180910390f35b6100a660048036038101906100a19190610381565b610161565b005b6100c260048036038101906100bd91906103bd565b610289565b6040516100d193929190610489565b60405180910390f35b6100f460048036038101906100ef9190610345565b6102a0565b005b6100fe6102ec565b60405161010b9190610445565b60405180910390f35b6000600160008375ffffffffffffffffffffffffffffffffffffffffffff1675ffffffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b7ff44c3bfbe5c9f6c6215c71d588688f093c5264ac95f9d7c15e06b899cc170a6160405160405180910390a17fdf94c109eaf09c6ae0f6f63ad3848d4f23f5baca53c076d8d6096833e0b6be5b826040516101bc9190610445565b60405180910390a1817f67b15b7b119723e10cda66d64abe2c5abbdad58a80c7fe87a84269870b479b7d60405160405180910390a27f8d191de11ad932288883c26fb31ec7f22554bcffe24580f671bf01766e38d05f81604051610220919061042a565b60405180910390a18075ffffffffffffffffffffffffffffffffffffffffffff167f52818dd789bafd2b091a2020832994407f938992a31a0084d80f0f7c1e26482460405160405180910390a2818160405161027d929190610460565b60405180910390a05050565b600080600085858592509250925093509350939050565b80600160008475ffffffffffffffffffffffffffffffffffffffffffff1675ffffffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055505050565b60005481565b600081359050610301816104fe565b92915050565b60008135905061031681610515565b92915050565b60006020828403121561032e57600080fd5b600061033c848285016102f2565b91505092915050565b6000806040838503121561035857600080fd5b6000610366858286016102f2565b925050602061037785828601610307565b9150509250929050565b6000806040838503121561039457600080fd5b60006103a285828601610307565b92505060206103b3858286016102f2565b9150509250929050565b6000806000606084860312156103d257600080fd5b60006103e086828701610307565b93505060206103f186828701610307565b925050604061040286828701610307565b9150509250925092565b610415816104c0565b82525050565b610424816104f4565b82525050565b600060208201905061043f600083018461040c565b92915050565b600060208201905061045a600083018461041b565b92915050565b6000604082019050610475600083018561041b565b610482602083018461040c565b9392505050565b600060608201905061049e600083018661041b565b6104ab602083018561041b565b6104b8604083018461041b565b949350505050565b60006104cb826104d2565b9050919050565b600075ffffffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b610507816104c0565b811461051257600080fd5b50565b61051e816104f4565b811461052957600080fd5b5056fea26469706673582212206960e5ea2cb996a02db763c74fb25590dc5b7ac8cddaa35a036d96cac121451e64736f6c637827302e382e342d646576656c6f702e323032322e382e32322b636f6d6d69742e61303164646338320058000000000000000000000000000000000000000000000000000000000000123400000000000000000000000000000000000000000000000000000000000012340000000000000000000000000000000000000000000000000000000000001234")
	
	// test contract code as deployed
	runtimeCode = common.Hex2Bytes("608060405234801561001057600080fd5b50600436106100575760003560e01c80630beac9b51461005c57806345c8ce261461008c578063498059b9146100a8578063e38fe5c9146100da578063f62fb0e9146100f6575b600080fd5b6100766004803603810190610071919061031c565b610114565b6040516100839190610445565b60405180910390f35b6100a660048036038101906100a19190610381565b610161565b005b6100c260048036038101906100bd91906103bd565b610289565b6040516100d193929190610489565b60405180910390f35b6100f460048036038101906100ef9190610345565b6102a0565b005b6100fe6102ec565b60405161010b9190610445565b60405180910390f35b6000600160008375ffffffffffffffffffffffffffffffffffffffffffff1675ffffffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b7ff44c3bfbe5c9f6c6215c71d588688f093c5264ac95f9d7c15e06b899cc170a6160405160405180910390a17fdf94c109eaf09c6ae0f6f63ad3848d4f23f5baca53c076d8d6096833e0b6be5b826040516101bc9190610445565b60405180910390a1817f67b15b7b119723e10cda66d64abe2c5abbdad58a80c7fe87a84269870b479b7d60405160405180910390a27f8d191de11ad932288883c26fb31ec7f22554bcffe24580f671bf01766e38d05f81604051610220919061042a565b60405180910390a18075ffffffffffffffffffffffffffffffffffffffffffff167f52818dd789bafd2b091a2020832994407f938992a31a0084d80f0f7c1e26482460405160405180910390a2818160405161027d929190610460565b60405180910390a05050565b600080600085858592509250925093509350939050565b80600160008475ffffffffffffffffffffffffffffffffffffffffffff1675ffffffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055505050565b60005481565b600081359050610301816104fe565b92915050565b60008135905061031681610515565b92915050565b60006020828403121561032e57600080fd5b600061033c848285016102f2565b91505092915050565b6000806040838503121561035857600080fd5b6000610366858286016102f2565b925050602061037785828601610307565b9150509250929050565b6000806040838503121561039457600080fd5b60006103a285828601610307565b92505060206103b3858286016102f2565b9150509250929050565b6000806000606084860312156103d257600080fd5b60006103e086828701610307565b93505060206103f186828701610307565b925050604061040286828701610307565b9150509250925092565b610415816104c0565b82525050565b610424816104f4565b82525050565b600060208201905061043f600083018461040c565b92915050565b600060208201905061045a600083018461041b565b92915050565b6000604082019050610475600083018561041b565b610482602083018461040c565b9392505050565b600060608201905061049e600083018661041b565b6104ab602083018561041b565b6104b8604083018461041b565b949350505050565b60006104cb826104d2565b9050919050565b600075ffffffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b610507816104c0565b811461051257600080fd5b50565b61051e816104f4565b811461052957600080fd5b5056fea26469706673582212206960e5ea2cb996a02db763c74fb25590dc5b7ac8cddaa35a036d96cac121451e64736f6c637827302e382e342d646576656c6f702e323032322e382e32322b636f6d6d69742e61303164646338320058")
	// contractSrc is predeploy on the following address in the genesis block.
	predeployedContractAddr, _ = common.HexToAddress("ce060000000000000000000000000000000000000314")
	
	// contractSrc is pre-deployed with the following address in the genesis block.
	predeployedContractWithAddress, _ = common.HexToAddress("ce632f8172fc6341a6f22dfe782f9b2ca2b728277efa")
	// holds the pre-deployed contract ABI
	predeployedContractABI = `[{"inputs":[{"internalType":"uint256","name":"ui_","type":"uint256"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[],"name":"E0","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"E1","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"","type":"uint256"}],"name":"E2","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"","type":"address"}],"name":"E3","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"","type":"address"}],"name":"E4","type":"event"},{"anonymous":true,"inputs":[{"indexed":false,"internalType":"uint256","name":"","type":"uint256"},{"indexed":false,"internalType":"address","name":"","type":"address"}],"name":"E5","type":"event"},{"inputs":[{"internalType":"address","name":"addr","type":"address"},{"internalType":"uint256","name":"value","type":"uint256"}],"name":"addToMap","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256","name":"b","type":"uint256"},{"internalType":"uint256","name":"c","type":"uint256"}],"name":"constFunc","outputs":[{"internalType":"uint256","name":"","type":"uint256"},{"internalType":"uint256","name":"","type":"uint256"},{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"ui_","type":"uint256"},{"internalType":"address","name":"addr_","type":"address"}],"name":"events","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"addr","type":"address"}],"name":"getFromMap","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"ui","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`
)

var (
	big0 = new(big.Int)
	big1 = big.NewInt(1)
)

// CodeAtTest tests the code for the pre-deployed contract.
func CodeAtTest(t *TestEnv) {
	code, err := t.Xcb.CodeAt(t.Ctx(), predeployedContractAddr, big0)
	if err != nil {
		t.Fatalf("Could not fetch code for predeployed contract: %v", err)
	}
	if bytes.Compare(runtimeCode, code) != 0 {
		t.Fatalf("Unexpected code, want %x, got %x", runtimeCode, code)
	}
}

// estimateEnergyTest fetches the estimated energy usage for a call to the events method.
func estimateEnergyTest(t *TestEnv) {
	var (
		address        = t.Vault.createAccount(t, big.NewInt(params.Core))
		contractABI, _ = abi.JSON(strings.NewReader(predeployedContractABI))
		intArg         = big.NewInt(rand.Int63())
	)

	payload, err := contractABI.Pack("events", intArg, address)
	if err != nil {
		t.Fatalf("Unable to prepare tx payload: %v", err)
	}
	msg := core.CallMsg{
		From: address,
		To:   &predeployedContractAddr,
		Data: payload,
	}
	estimated, err := t.Xcb.EstimateEnergy(t.Ctx(), msg)
	if err != nil {
		t.Fatalf("Could not estimate energy: %v", err)
	}

	// send the actual tx and test energy usage
	txenergy := estimated + 100000
	rawTx := types.NewTransaction(0, *msg.To, msg.Value, txenergy, big.NewInt(32*params.Nucle), msg.Data)
	tx, err := t.Vault.signTransaction(address, rawTx)
	if err != nil {
		t.Fatalf("Could not sign transaction: %v", err)
	}

	if err := t.Xcb.SendTransaction(t.Ctx(), tx); err != nil {
		t.Fatalf("Could not send tx: %v", err)
	}

	receipt, err := waitForTxConfirmations(t, tx.Hash(), 1)
	if err != nil {
		t.Fatalf("Could not wait for confirmations: %v", err)
	}

	// test lower bound
	if estimated < receipt.EnergyUsed {
		t.Fatalf("Estimated energy too low, want %d >= %d", estimated, receipt.EnergyUsed)
	}
	// test upper bound
	if receipt.EnergyUsed+5000 < estimated {
		t.Fatalf("Estimated energy too high, estimated: %d, used: %d", estimated, receipt.EnergyUsed)
	}
}

// balanceAndNonceAtTest creates a new account and transfers funds to it.
// It then tests if the balance and nonce of the sender and receiver
// address are updated correct.
func balanceAndNonceAtTest(t *TestEnv) {
	var (
		sourceAddr  = t.Vault.createAccount(t, big.NewInt(params.Core))
		sourceNonce = uint64(0)
		targetAddr  = t.Vault.createAccount(t, nil)
	)

	// Get current balance
	sourceAddressBalanceBefore, err := t.Xcb.BalanceAt(t.Ctx(), sourceAddr, nil)
	if err != nil {
		t.Fatalf("Unable to retrieve balance: %v", err)
	}

	expected := big.NewInt(params.Core)
	if sourceAddressBalanceBefore.Cmp(expected) != 0 {
		t.Errorf("Expected balance %d, got %d", expected, sourceAddressBalanceBefore)
	}

	nonceBefore, err := t.Xcb.NonceAt(t.Ctx(), sourceAddr, nil)
	if err != nil {
		t.Fatalf("Unable to determine nonce: %v", err)
	}
	if nonceBefore != sourceNonce {
		t.Fatalf("Invalid nonce, want %d, got %d", sourceNonce, nonceBefore)
	}

	// send 1234 wei to target account and verify balances and nonces are updated
	var (
		amount   = big.NewInt(1234)
		energyLimit = uint64(50000)
	)
	rawTx := types.NewTransaction(sourceNonce, targetAddr, amount, energyLimit, energyPrice, nil)
	valueTx, err := t.Vault.signTransaction(sourceAddr, rawTx)
	if err != nil {
		t.Fatalf("Unable to sign value tx: %v", err)
	}
	sourceNonce++

	t.Logf("BalanceAt: send %d wei from 0x%x to 0x%x in 0x%x", valueTx.Value(), sourceAddr, targetAddr, valueTx.Hash())
	if err := t.Xcb.SendTransaction(t.Ctx(), valueTx); err != nil {
		t.Fatalf("Unable to send transaction: %v", err)
	}

	var receipt *types.Receipt
	for {
		receipt, err = t.Xcb.TransactionReceipt(t.Ctx(), valueTx.Hash())
		if receipt != nil {
			break
		}
		if err != core.NotFound {
			t.Fatalf("Could not fetch receipt for 0x%x: %v", valueTx.Hash(), err)
		}
		time.Sleep(time.Second)
	}

	// ensure balances have been updated
	accountBalanceAfter, err := t.Xcb.BalanceAt(t.Ctx(), sourceAddr, nil)
	if err != nil {
		t.Fatalf("Unable to retrieve balance: %v", err)
	}
	balanceTargetAccountAfter, err := t.Xcb.BalanceAt(t.Ctx(), targetAddr, nil)
	if err != nil {
		t.Fatalf("Unable to retrieve balance: %v", err)
	}

	// expected balance is previous balance - tx amount - tx fee (energyUsed * energyPrice)
	exp := new(big.Int).Set(sourceAddressBalanceBefore)
	exp.Sub(exp, amount)
	exp.Sub(exp, new(big.Int).Mul(big.NewInt(int64(receipt.EnergyUsed)), valueTx.EnergyPrice()))

	if exp.Cmp(accountBalanceAfter) != 0 {
		t.Errorf("Expected sender account to have a balance of %d, got %d", exp, accountBalanceAfter)
	}
	if balanceTargetAccountAfter.Cmp(amount) != 0 {
		t.Errorf("Expected new account to have a balance of %d, got %d", valueTx.Value(), balanceTargetAccountAfter)
	}

	// ensure nonce is incremented by 1
	nonceAfter, err := t.Xcb.NonceAt(t.Ctx(), sourceAddr, nil)
	if err != nil {
		t.Fatalf("Unable to determine nonce: %v", err)
	}
	expectedNonce := nonceBefore + 1
	if expectedNonce != nonceAfter {
		t.Fatalf("Invalid nonce, want %d, got %d", expectedNonce, nonceAfter)
	}
}

// genesisByHash fetches the known genesis header and compares
// it against the genesis file to determine if block fields are
// returned correct.
func genesisHeaderByHashTest(t *TestEnv) {
	gblock := loadGenesis()

	headerByHash, err := t.Xcb.HeaderByHash(t.Ctx(), gblock.Hash())
	if err != nil {
		t.Fatalf("Unable to fetch block %x: %v", gblock.Hash(), err)
	}
	if d := diff(gblock.Header(), headerByHash); d != "" {
		t.Fatal("genesis header reported by node differs from expected header:\n", d)
	}
}

// headerByNumberTest fetched the known genesis header and compares
// it against the genesis file to determine if block fields are
// returned correct.
func genesisHeaderByNumberTest(t *TestEnv) {
	gblock := loadGenesis()

	headerByNum, err := t.Xcb.HeaderByNumber(t.Ctx(), big0)
	if err != nil {
		t.Fatalf("Unable to fetch genesis block: %v", err)
	}
	if d := diff(gblock.Header(), headerByNum); d != "" {
		t.Fatal("genesis header reported by node differs from expected header:\n", d)
	}
}

// genesisBlockByHashTest fetched the known genesis block and compares it against
// the genesis file to determine if block fields are returned correct.
func genesisBlockByHashTest(t *TestEnv) {
	gblock := loadGenesis()

	blockByHash, err := t.Xcb.BlockByHash(t.Ctx(), gblock.Hash())
	if err != nil {
		t.Fatalf("Unable to fetch block %x: %v", gblock.Hash(), err)
	}
	if d := diff(gblock.Header(), blockByHash.Header()); d != "" {
		t.Fatal("genesis header reported by node differs from expected header:\n", d)
	}
}

// genesisBlockByNumberTest retrieves block 0 since that is the only block
// that is known through the genesis.json file and tests if block
// fields matches the fields defined in the genesis file.
func genesisBlockByNumberTest(t *TestEnv) {
	gblock := loadGenesis()

	blockByNum, err := t.Xcb.BlockByNumber(t.Ctx(), big0)
	if err != nil {
		t.Fatalf("Unable to fetch genesis block: %v", err)
	}
	if d := diff(gblock.Header(), blockByNum.Header()); d != "" {
		t.Fatal("genesis header reported by node differs from expected header:\n", d)
	}
}

// canonicalChainTest loops over 10 blocks and does some basic validations
// to ensure the chain form a valid canonical chain and resources like uncles,
// transactions and receipts can be fetched and provide a consistent view.
func canonicalChainTest(t *TestEnv) {
	// wait a bit so there is actually a chain with enough height
	for {
		latestBlock, err := t.Xcb.BlockByNumber(t.Ctx(), nil)
		if err != nil {
			t.Fatalf("Unable to fetch latest block")
		}
		if latestBlock.NumberU64() >= 20 {
			break
		}
		time.Sleep(time.Second)
	}

	var childBlock *types.Block
	for i := 10; i >= 0; i-- {
		block, err := t.Xcb.BlockByNumber(t.Ctx(), big.NewInt(int64(i)))
		if err != nil {
			t.Fatalf("Unable to fetch block #%d", i)
		}
		if childBlock != nil {
			if childBlock.ParentHash() != block.Hash() {
				t.Errorf("Canonical chain broken on %d-%d / %x-%x", block.NumberU64(), childBlock.NumberU64(), block.Hash(), childBlock.Hash())
			}
		}

		// try to fetch all txs and receipts and do some basic validation on them
		// to check if the fetched chain is consistent.
		for _, tx := range block.Transactions() {
			fetchedTx, _, err := t.Xcb.TransactionByHash(t.Ctx(), tx.Hash())
			if err != nil {
				t.Fatalf("Unable to fetch transaction %x from block %x: %v", tx.Hash(), block.Hash(), err)
			}
			if fetchedTx == nil {
				t.Fatalf("Transaction %x could not be found but was included in block %x", tx.Hash(), block.Hash())
			}
			receipt, err := t.Xcb.TransactionReceipt(t.Ctx(), fetchedTx.Hash())
			if err != nil {
				t.Fatalf("Unable to fetch receipt for %x from block %x: %v", fetchedTx.Hash(), block.Hash(), err)
			}
			if receipt == nil {
				t.Fatalf("Receipt for %x could not be found but was included in block %x", fetchedTx.Hash(), block.Hash())
			}
			if receipt.TxHash != fetchedTx.Hash() {
				t.Fatalf("Receipt has an invalid tx, expected %x, got %x", fetchedTx.Hash(), receipt.TxHash)
			}
		}

		// make sure all uncles can be fetched
		for _, uncle := range block.Uncles() {
			uBlock, err := t.Xcb.HeaderByHash(t.Ctx(), uncle.Hash())
			if err != nil {
				t.Fatalf("Unable to fetch uncle block: %v", err)
			}
			if uBlock == nil {
				t.Logf("Could not fetch uncle block %x", uncle.Hash())
			}
		}

		childBlock = block
	}
}

// deployContractTest deploys `contractSrc` and tests if the code and state
// on the contract address contain the expected values (as set in the ctor).
func deployContractTest(t *TestEnv) {
	var (
		address = t.Vault.createAccount(t, big.NewInt(params.Core))
		nonce   = uint64(0)

		expectedContractAddress = crypto.CreateAddress(address, nonce)
		energyLimit                = uint64(1200000)
	)

	rawTx := types.NewContractCreation(nonce, big0, energyLimit, energyPrice, deployCode)
	deployTx, err := t.Vault.signTransaction(address, rawTx)
	if err != nil {
		t.Fatalf("Unable to sign deploy tx: %v", err)
	}

	// deploy contract
	if err := t.Xcb.SendTransaction(t.Ctx(), deployTx); err != nil {
		t.Fatalf("Unable to send transaction: %v", err)
	}

	t.Logf("Deploy transaction: 0x%x", deployTx.Hash())

	// fetch transaction receipt for contract address
	var contractAddress common.Address
	receipt, err := waitForTxConfirmations(t, deployTx.Hash(), 5)
	if err != nil {
		t.Fatalf("Unable to retrieve receipt: %v", err)
	}

	// ensure receipt has the expected address
	if expectedContractAddress != receipt.ContractAddress {
		t.Fatalf("Contract deploy on different address, expected %x, got %x", expectedContractAddress, contractAddress)
	}

	// test deployed code matches runtime code
	code, err := t.Xcb.CodeAt(t.Ctx(), receipt.ContractAddress, nil)
	if err != nil {
		t.Fatalf("Unable to fetch contract code: %v", err)
	}
	if bytes.Compare(runtimeCode, code) != 0 {
		t.Errorf("Deployed code doesn't match, expected %x, got %x", runtimeCode, code)
	}

	// test contract state, pos 0 must be 4660
	value, err := t.Xcb.StorageAt(t.Ctx(), receipt.ContractAddress, common.Hash{}, nil)
	if err == nil {
		v := new(big.Int).SetBytes(value)
		if v.Uint64() != 4660 {
			t.Errorf("Unexpected value on %x:0x01, expected 4660, got %d", receipt.ContractAddress, v)
		}
	} else {
		t.Errorf("Unable to retrieve storage pos 0x01 on address %x: %v", contractAddress, err)
	}

	// test contract state, map on pos 1 with key myAccount must be 4660
	storageKey := make([]byte, 64)
	copy(storageKey[10:32], address.Bytes())
	storageKey[63] = 1
	storageKey = crypto.SHA3(storageKey)

	value, err = t.Xcb.StorageAt(t.Ctx(), receipt.ContractAddress, common.BytesToHash(storageKey), nil)
	if err == nil {
		v := new(big.Int).SetBytes(value)
		if v.Uint64() != 4660 {
			t.Errorf("Unexpected value in map, expected 4660, got %d", v)
		}
	} else {
		t.Fatalf("Unable to retrieve value in map: %v", err)
	}
}

// deployContractOutOfEnergyTest tries to deploy `contractSrc` with insufficient energy. It
// checks the receipts reflects the "out of energy" event and code / state isn't created in
// the contract address.
func deployContractOutOfEnergyTest(t *TestEnv) {
	var (
		address         = t.Vault.createAccount(t, big.NewInt(params.Core))
		nonce           = uint64(0)
		contractAddress = crypto.CreateAddress(address, nonce)
		energyLimit        = uint64(110000) // insufficient energy
	)
	t.Logf("calculated contract address: %x", contractAddress)

	// Deploy the contract.
	rawTx := types.NewContractCreation(nonce, big0, energyLimit, energyPrice, deployCode)
	deployTx, err := t.Vault.signTransaction(address, rawTx)
	if err != nil {
		t.Fatalf("unable to sign deploy tx: %v", err)
	}
	t.Logf("out of energy tx: %x", deployTx.Hash())
	if err := t.Xcb.SendTransaction(t.Ctx(), deployTx); err != nil {
		t.Fatalf("unable to send transaction: %v", err)
	}

	// Wait for the transaction receipt.
	receipt, err := waitForTxConfirmations(t, deployTx.Hash(), 5)
	if err != nil {
		t.Fatalf("unable to fetch tx receipt: %v", err)
	}
	// Check receipt fields.
	if receipt.Status != types.ReceiptStatusFailed {
		t.Errorf("receipt has status %d, want %d", receipt.Status, types.ReceiptStatusFailed)
	}
	if receipt.EnergyUsed != energyLimit {
		t.Errorf("receipt has energyUsed %d, want %d", receipt.EnergyUsed, energyLimit)
	}
	if receipt.ContractAddress != contractAddress {
		t.Errorf("receipt has contract address %x, want %x", receipt.ContractAddress, contractAddress)
	}
	if receipt.BlockHash == (common.Hash{}) {
		t.Errorf("receipt has empty block hash", receipt.BlockHash)
	}
	// Check that nothing is deployed at the contract address.
	code, err := t.Xcb.CodeAt(t.Ctx(), contractAddress, nil)
	if err != nil {
		t.Fatalf("unable to fetch code: %v", err)
	}
	if len(code) != 0 {
		t.Errorf("expected no code deployed but got %x", code)
	}
}

// receiptTest tests whether the created receipt is correct by calling the `events` method
// on the pre-deployed contract.
func receiptTest(t *TestEnv) {
	var (
		contractABI, _ = abi.JSON(strings.NewReader(predeployedContractABI))
		address        = t.Vault.createAccount(t, big.NewInt(params.Core))
		nonce          = uint64(0)

		intArg = big.NewInt(rand.Int63())
	)

	payload, err := contractABI.Pack("events", intArg, address)
	if err != nil {
		t.Fatalf("Unable to prepare tx payload: %v", err)
	}

	rawTx := types.NewTransaction(nonce, predeployedContractAddr, big0, 500000, energyPrice, payload)
	tx, err := t.Vault.signTransaction(address, rawTx)
	if err != nil {
		t.Fatalf("Unable to sign deploy tx: %v", err)
	}

	if err := t.Xcb.SendTransaction(t.Ctx(), tx); err != nil {
		t.Fatalf("Unable to send transaction: %v", err)
	}

	// wait for transaction
	receipt, err := waitForTxConfirmations(t, tx.Hash(), 0)
	if err != nil {
		t.Fatalf("Unable to retrieve tx receipt: %v", err)
	}
	// validate receipt fields
	if receipt.TxHash != tx.Hash() {
		t.Errorf("Receipt contains invalid tx hash, want %x, got %x", tx.Hash(), receipt.TxHash)
	}
	if receipt.ContractAddress != (common.Address{}) {
		t.Errorf("Receipt contains invalid contract address, want empty address got %x", receipt.ContractAddress)
	}
	bloom := types.CreateBloom(types.Receipts{receipt})
	if receipt.Bloom != bloom {
		t.Errorf("Receipt contains invalid bloom, want %x, got %x", bloom, receipt.Bloom)
	}

	var (
		intArgBytes  = common.LeftPadBytes(intArg.Bytes(), 32)
		addrArgBytes = common.LeftPadBytes(address.Bytes(), 32)
	)

	if len(receipt.Logs) != 6 {
		t.Fatalf("Want 6 logs, got %d", len(receipt.Logs))
	}

	validateLog(t, tx, *receipt.Logs[0], predeployedContractAddr, receipt.Logs[0].Index+0, contractABI.Events["E0"], nil)
	validateLog(t, tx, *receipt.Logs[1], predeployedContractAddr, receipt.Logs[0].Index+1, contractABI.Events["E1"], intArgBytes)
	validateLog(t, tx, *receipt.Logs[2], predeployedContractAddr, receipt.Logs[0].Index+2, contractABI.Events["E2"], intArgBytes)
	validateLog(t, tx, *receipt.Logs[3], predeployedContractAddr, receipt.Logs[0].Index+3, contractABI.Events["E3"], addrArgBytes)
	validateLog(t, tx, *receipt.Logs[4], predeployedContractAddr, receipt.Logs[0].Index+4, contractABI.Events["E4"], addrArgBytes)
	validateLog(t, tx, *receipt.Logs[5], predeployedContractAddr, receipt.Logs[0].Index+5, contractABI.Events["E5"], intArgBytes, addrArgBytes)
}

// validateLog is a helper method that tests if the given set of logs are valid when the events method on the
// standard contract is called with argData.
func validateLog(t *TestEnv, tx *types.Transaction, log types.Log, contractAddress common.Address, index uint, ev abi.Event, argData ...[]byte) {
	if log.Address != contractAddress {
		t.Errorf("Log[%d] contains invalid address, want 0x%x, got 0x%x [tx=0x%x]", index, contractAddress, log.Address, tx.Hash())
	}
	if log.TxHash != tx.Hash() {
		t.Errorf("Log[%d] contains invalid hash, want 0x%x, got 0x%x [tx=0x%x]", index, tx.Hash(), log.TxHash, tx.Hash())
	}
	if log.Index != index {
		t.Errorf("Log[%d] has invalid index, want %d, got %d [tx=0x%x]", index, index, log.Index, tx.Hash())
	}

	// assemble expected topics and log data
	var (
		topics []common.Hash
		data   []byte
	)
	if !ev.Anonymous {
		topics = append(topics, ev.ID)
	}
	for i, arg := range ev.Inputs {
		if arg.Indexed {
			topics = append(topics, common.BytesToHash(argData[i]))
		} else {
			data = append(data, argData[i]...)
		}
	}

	if len(log.Topics) != len(topics) {
		t.Errorf("Log[%d] contains invalid number of topics, want %d, got %d [tx=0x%x]", index, len(topics), len(log.Topics), tx.Hash())
	} else {
		for i, topic := range topics {
			if topics[i] != topic {
				t.Errorf("Log[%d] contains invalid topic, want 0x%x, got 0x%x [tx=0x%x]", index, topics[i], topic, tx.Hash())
			}
		}
	}
	if !bytes.Equal(log.Data, data) {
		t.Errorf("Log[%d] contains invalid data, want 0x%x, got 0x%x [tx=0x%x]", index, data, log.Data, tx.Hash())
	}
}

// syncProgressTest only tests if this function is supported by the node.
func syncProgressTest(t *TestEnv) {
	_, err := t.Xcb.SyncProgress(t.Ctx())
	if err != nil {
		t.Fatalf("Unable to determine sync progress: %v", err)
	}
}

// transactionInBlockTest will wait for a new block with transaction
// and retrieves transaction details by block hash and position.
func transactionInBlockTest(t *TestEnv) {
	var (
		key         = t.Vault.createAccount(t, big.NewInt(params.Core))
		nonce       = uint64(0)
		blockNumber = new(big.Int)
	)

	for {
		blockNumber.Add(blockNumber, big1)

		block, err := t.Xcb.BlockByNumber(t.Ctx(), blockNumber)
		if err == core.NotFound { // end of chain
			rawTx := types.NewTransaction(nonce, predeployedVaultAddr, big1, 100000, energyPrice, nil)
			nonce++

			tx, err := t.Vault.signTransaction(key, rawTx)
			if err != nil {
				t.Fatalf("Unable to sign deploy tx: %v", err)
			}
			if err = t.Xcb.SendTransaction(t.Ctx(), tx); err != nil {
				t.Fatalf("Unable to send transaction: %v", err)
			}
			time.Sleep(time.Second)
			continue
		}
		if err != nil {
			t.Fatalf("Unable to fetch latest block: %v", err)
		}
		if len(block.Transactions()) == 0 {
			continue
		}
		for i := 0; i < len(block.Transactions()); i++ {
			_, err := t.Xcb.TransactionInBlock(t.Ctx(), block.Hash(), uint(i))
			if err != nil {
				t.Fatalf("Unable to fetch transaction by block hash and index: %v", err)
			}
		}
		return
	}
}

// transactionInBlockSubscriptionTest will wait for a new block with transaction
// and retrieves transaction details by block hash and position.
func transactionInBlockSubscriptionTest(t *TestEnv) {
	var heads = make(chan *types.Header, 100)

	sub, err := t.Xcb.SubscribeNewHead(t.Ctx(), heads)
	if err != nil {
		t.Fatalf("Unable to subscribe to new heads: %v", err)
	}

	key := t.Vault.createAccount(t, big.NewInt(params.Core))
	for i := 0; i < 5; i++ {
		rawTx := types.NewTransaction(uint64(i), predeployedVaultAddr, big1, 100000, energyPrice, nil)
		tx, err := t.Vault.signTransaction(key, rawTx)
		if err != nil {
			t.Fatalf("Unable to sign deploy tx: %v", err)
		}
		if err = t.Xcb.SendTransaction(t.Ctx(), tx); err != nil {
			t.Fatalf("Unable to send transaction: %v", err)
		}
	}

	// wait until transaction
	defer sub.Unsubscribe()
	for {
		head := <-heads

		block, err := t.Xcb.BlockByHash(t.Ctx(), head.Hash())
		if err != nil {
			t.Fatalf("Unable to retrieve block %x: %v", head.Hash(), err)
		}
		if len(block.Transactions()) == 0 {
			continue
		}
		for i := 0; i < len(block.Transactions()); i++ {
			_, err = t.Xcb.TransactionInBlock(t.Ctx(), head.Hash(), uint(i))
			if err != nil {
				t.Fatalf("Unable to fetch transaction by block hash and index: %v", err)
			}
		}
		return
	}
}

// newHeadSubscriptionTest tests whether
func newHeadSubscriptionTest(t *TestEnv) {
	var (
		heads = make(chan *types.Header)
	)

	sub, err := t.Xcb.SubscribeNewHead(t.Ctx(), heads)
	if err != nil {
		t.Fatalf("Unable to subscribe to new heads: %v", err)
	}

	defer sub.Unsubscribe()
	for i := 0; i < 10; i++ {
		select {
		case newHead := <-heads:
			header, err := t.Xcb.HeaderByHash(t.Ctx(), newHead.Hash())
			if err != nil {
				t.Fatalf("Unable to fetch header: %v", err)
			}
			if header == nil {
				t.Fatalf("Unable to fetch header %s", newHead.Hash())
			}
		case err := <-sub.Err():
			t.Fatalf("Received errors: %v", err)
		}
	}
}

func logSubscriptionTest(t *TestEnv) {
	var (
		criteria = core.FilterQuery{
			Addresses: []common.Address{predeployedContractAddr},
			Topics:    [][]common.Hash{},
		}
		logs = make(chan types.Log)
	)

	sub, err := t.Xcb.SubscribeFilterLogs(t.Ctx(), criteria, logs)
	if err != nil {
		t.Fatalf("Unable to create log subscription: %v", err)
	}
	defer sub.Unsubscribe()

	var (
		contractABI, _ = abi.JSON(strings.NewReader(predeployedContractABI))
		address        = t.Vault.createAccount(t, big.NewInt(params.Core))
		nonce          = uint64(0)

		arg0 = big.NewInt(rand.Int63())
		arg1 = address
	)

	payload, _ := contractABI.Pack("events", arg0, arg1)
	rawTx := types.NewTransaction(nonce, predeployedContractAddr, big0, 500000, energyPrice, payload)
	tx, err := t.Vault.signTransaction(address, rawTx)
	if err != nil {
		t.Fatalf("Unable to sign deploy tx: %v", err)
	}

	if err = t.Xcb.SendTransaction(t.Ctx(), tx); err != nil {
		t.Fatalf("Unable to send transaction: %v", err)
	}

	t.Logf("Wait for logs generated for transaction: %x", tx.Hash())
	var (
		expectedLogs = 6
		currentLogs  = 0
		fetchedLogs  []types.Log
		deadline     = time.NewTimer(30 * time.Second)
	)

	// ensure we receive all logs that are generated by our transaction.
	// log fields are in depth verified in another test.
	for len(fetchedLogs) < expectedLogs {
		select {
		case log := <-logs:
			// other tests also send transaction to the predeployed
			// contract ensure these logs are from "our" transaction.
			if log.TxHash != tx.Hash() {
				continue
			}
			fetchedLogs = append(fetchedLogs, log)
		case err := <-sub.Err():
			t.Fatalf("Log subscription returned error: %v", err)
		case <-deadline.C:
			t.Fatalf("Only received %d/%d logs", currentLogs, expectedLogs)
		}
	}

	validatePredeployContractLogs(t, tx, fetchedLogs, arg0, arg1)
}

// validatePredeployContractLogs tests wether the given logs are expected when
// the event function was called on the predeployed test contract was called
// with the given args. The event function raises the following events:
// event E0();
// event E1(uint);
// event E2(uint indexed);
// event E3(address);
// event E4(address indexed);
// event E5(uint, address) anonymous;
func validatePredeployContractLogs(t *TestEnv, tx *types.Transaction, logs []types.Log, intArg *big.Int, addrArg common.Address) {
	if len(logs) != 6 {
		t.Fatalf("Unexpected log count, want 6, got %d", len(logs))
	}

	var (
		contractABI, _ = abi.JSON(strings.NewReader(predeployedContractABI))
		intArgBytes    = common.LeftPadBytes(intArg.Bytes(), 32)
		addrArgBytes   = common.LeftPadBytes(addrArg.Bytes(), 32)
	)

	validateLog(t, tx, logs[0], predeployedContractAddr, logs[0].Index+0, contractABI.Events["E0"], nil)
	validateLog(t, tx, logs[1], predeployedContractAddr, logs[0].Index+1, contractABI.Events["E1"], intArgBytes)
	validateLog(t, tx, logs[2], predeployedContractAddr, logs[0].Index+2, contractABI.Events["E2"], intArgBytes)
	validateLog(t, tx, logs[3], predeployedContractAddr, logs[0].Index+3, contractABI.Events["E3"], addrArgBytes)
	validateLog(t, tx, logs[4], predeployedContractAddr, logs[0].Index+4, contractABI.Events["E4"], addrArgBytes)
	validateLog(t, tx, logs[5], predeployedContractAddr, logs[0].Index+5, contractABI.Events["E5"], intArgBytes, addrArgBytes)
}

func transactionCountTest(t *TestEnv) {
	var (
		key = t.Vault.createAccount(t, big.NewInt(params.Core))
	)

	for i := 0; i < 60; i++ {
		rawTx := types.NewTransaction(uint64(i), predeployedVaultAddr, big1, 100000, energyPrice, nil)
		tx, err := t.Vault.signTransaction(key, rawTx)
		if err != nil {
			t.Fatalf("Unable to sign deploy tx: %v", err)
		}

		if err = t.Xcb.SendTransaction(t.Ctx(), tx); err != nil {
			t.Fatalf("Unable to send transaction: %v", err)
		}
		block, err := t.Xcb.BlockByNumber(t.Ctx(), nil)
		if err != nil {
			t.Fatalf("Unable to retrieve latest block: %v", err)
		}

		if len(block.Transactions()) > 0 {
			count, err := t.Xcb.TransactionCount(t.Ctx(), block.Hash())
			if err != nil {
				t.Fatalf("Unable to retrieve block transaction count: %v", err)
			}
			if count != uint(len(block.Transactions())) {
				t.Fatalf("Invalid block tx count, want %d, got %d", len(block.Transactions()), count)
			}
			return
		}

		time.Sleep(time.Second)
	}
}

// TransactionReceiptTest sends a transaction and tests the receipt fields.
func TransactionReceiptTest(t *TestEnv) {
	var (
		key = t.Vault.createAccount(t, big.NewInt(params.Core))
	)

	rawTx := types.NewTransaction(uint64(0), common.Address{}, big1, 100000, energyPrice, nil)
	tx, err := t.Vault.signTransaction(key, rawTx)
	if err != nil {
		t.Fatalf("Unable to sign deploy tx: %v", err)
	}

	if err = t.Xcb.SendTransaction(t.Ctx(), tx); err != nil {
		t.Fatalf("Unable to send transaction: %v", err)
	}

	for i := 0; i < 60; i++ {
		receipt, err := t.Xcb.TransactionReceipt(t.Ctx(), tx.Hash())
		if err == core.NotFound {
			time.Sleep(time.Second)
			continue
		}

		if err != nil {
			t.Errorf("Unable to fetch receipt: %v", err)
		}
		if receipt.TxHash != tx.Hash() {
			t.Errorf("Receipt [tx=%x] contains invalid tx hash, want %x, got %x", tx.Hash(), receipt.TxHash)
		}
		if receipt.ContractAddress != (common.Address{}) {
			t.Errorf("Receipt [tx=%x] contains invalid contract address, expected empty address but got %x", tx.Hash(), receipt.ContractAddress)
		}
		if receipt.Bloom.Big().Cmp(big0) != 0 {
			t.Errorf("Receipt [tx=%x] bloom not empty, %x", tx.Hash(), receipt.Bloom)
		}
		if receipt.EnergyUsed != params.TxEnergy {
			t.Errorf("Receipt [tx=%x] has invalid energy used, want %d, got %d", tx.Hash(), params.TxEnergy, receipt.EnergyUsed)
		}
		if len(receipt.Logs) != 0 {
			t.Errorf("Receipt [tx=%x] should not contain logs but got %d logs", tx.Hash(), len(receipt.Logs))
		}
		return
	}
}
