package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/core-coin/go-core/v2"
	"github.com/core-coin/go-core/v2/accounts/abi"
	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/core/types"
	"github.com/core-coin/go-core/v2/crypto"
)

var (
	// This is the account that sends vault funding transactions.
	vaultAccountAddr, _ = common.HexToAddress("cb74db416ff2f9c53dabaf34f81142db30350ea7b144")
	vaultKey, _      = crypto.UnmarshalPrivateKeyHex("8ba306058fe6fda07f3db9edbccef75ae4171081fe47cca47ae2bbdc8dd423f1e192550f02a7b7e25664bb9248d88265c03fa9f4963a65c66e")
	// Address of the vault in genesis.
	predeployedVaultAddr, _ = common.HexToAddress("cb850000000000000000000000000000000000000315")
	// Number of blocks to wait before funding tx is considered valid.
	vaultTxConfirmationCount = uint64(5)
)

// vault creates accounts for testing and funds them. An instance of the vault contract is
// deployed in the genesis block. When creating a new account using createAccount, the
// account is funded by sending a transaction to this contract.
//
// The purpose of the vault is allowing tests to run concurrently without worrying about
// nonce assignment and unexpected balance changes.
type vault struct {
	mu sync.Mutex
	// This tracks the account nonce of the vault account.
	nonce uint64
	// Created accounts are tracked in this map.
	accounts map[common.Address]*crypto.PrivateKey
}

func newVault() *vault {
	return &vault{
		accounts: make(map[common.Address]*crypto.PrivateKey),
	}
}

// generateKey creates a new account key and stores it.
func (v *vault) generateKey() common.Address {
	key, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		panic(fmt.Errorf("can't generate account key: %v", err))
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.accounts[key.Address()] = key
	return key.Address()
}

// findKey returns the private key for an address.
func (v *vault) findKey(addr common.Address) *crypto.PrivateKey {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.accounts[addr]
}

// signTransaction signs the given transaction with the test account and returns it.
func (v *vault) signTransaction(sender common.Address, tx *types.Transaction) (*types.Transaction, error) {
	key := v.findKey(sender)
	if key == nil {
		return nil, fmt.Errorf("sender account %v not in vault", sender)
	}
	signer := types.NewNucleusSigner(networkID)
	return types.SignTx(tx, signer, key)
}

// createAndFundAccount creates a new account that is funded from the vault contract.
// It will panic when the account could not be created and funded.
func (v *vault) createAccountWithSubscription(t *TestEnv, amount *big.Int) common.Address {
	if amount == nil {
		amount = new(big.Int)
	}
	address := v.generateKey()

	// setup subscriptions
	var (
		headsSub core.Subscription
		heads    = make(chan *types.Header)
		logsSub  core.Subscription
		logs     = make(chan types.Log)
		vault, _ = abi.JSON(strings.NewReader(predeployedVaultABI))
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// listen for new heads
	headsSub, err := t.Xcb.SubscribeNewHead(ctx, heads)
	if err != nil {
		t.Fatal("could not create new head subscription:", err)
	}
	defer headsSub.Unsubscribe()

	// set up the log event subscription
	eventTopic := vault.Events["Send"].ID
	addressTopic := common.BytesToHash(common.LeftPadBytes(address[:], 32))
	q := core.FilterQuery{
		Addresses: []common.Address{predeployedVaultAddr},
		Topics:    [][]common.Hash{{eventTopic}, {addressTopic}},
	}
	logsSub, err = t.Xcb.SubscribeFilterLogs(ctx, q, logs)
	if err != nil {
		t.Fatal("could not create log filter subscription:", err)
	}
	defer logsSub.Unsubscribe()

	// order the vault to send some xcb's
	tx := v.makeFundingTx(t, address, amount)
	if err := t.Xcb.SendTransaction(ctx, tx); err != nil {
		t.Fatalf("unable to send funding transaction: %v", err)
	}

	// wait for confirmed log
	var (
		latestHeader *types.Header
		receivedLog  *types.Log
		timeout      = time.NewTimer(120 * time.Second)
	)
	for {
		select {
		case head := <-heads:
			latestHeader = head
		case log := <-logs:
			if !log.Removed {
				receivedLog = &log
			} else if log.Removed && receivedLog != nil && receivedLog.BlockHash == log.BlockHash {
				// chain reorg!
				receivedLog = nil
			}
		case err := <-headsSub.Err():
			t.Fatalf("could not fund new account: %v", err)
		case err := <-logsSub.Err():
			t.Fatalf("could not fund new account: %v", err)
		case <-timeout.C:
			t.Fatal("could not fund new account: timeout")
		}

		if latestHeader != nil && receivedLog != nil {
			if receivedLog.BlockNumber+vaultTxConfirmationCount <= latestHeader.Number.Uint64() {
				return address
			}
		}
	}

	return address
}

// createAccount creates a new account that is funded from the vault contract.
// It will panic when the account could not be created and funded.
func (v *vault) createAccount(t *TestEnv, amount *big.Int) common.Address {
	if amount == nil {
		amount = new(big.Int)
	}
	address := v.generateKey()

	// order the vault to send some xcb's
	tx := v.makeFundingTx(t, address, amount)
	if err := t.Xcb.SendTransaction(t.Ctx(), tx); err != nil {
		t.Fatalf("unable to send funding transaction: %v", err)
	}

	txBlock, err := t.Xcb.BlockNumber(t.Ctx())
	if err != nil {
		t.Fatalf("can't get block number:", err)
	}

	// wait for vaultTxConfirmationCount confirmation by checking the balance vaultTxConfirmationCount blocks back.
	// createAndFundAccountWithSubscription for a better solution using logs
	for i := uint64(0); i < vaultTxConfirmationCount*12; i++ {
		number, err := t.Xcb.BlockNumber(t.Ctx())
		if err != nil {
			t.Fatalf("can't get block number:", err)
		}
		if number > txBlock+vaultTxConfirmationCount {
			checkBlock := number - vaultTxConfirmationCount
			balance, err := t.Xcb.BalanceAt(t.Ctx(), address, new(big.Int).SetUint64(checkBlock))
			if err != nil {
				panic(err)
			}
			if balance.Cmp(amount) >= 0 {
				return address
			}
		}
		time.Sleep(time.Second)
	}
	panic(fmt.Sprintf("could not fund account %v in transaction %v", address, tx.Hash()))
}

func (v *vault) makeFundingTx(t *TestEnv, recipient common.Address, amount *big.Int) *types.Transaction {
	vault, _ := abi.JSON(strings.NewReader(predeployedVaultABI))
	payload, err := vault.Pack("sendSome", recipient, amount)
	if err != nil {
		t.Fatalf("can't pack pack vault tx input: %v", err)
	}
	var (
		nonce    = v.nextNonce()
		energyLimit = uint64(75000)
		txAmount = new(big.Int)
	)
	tx := types.NewTransaction(nonce, predeployedVaultAddr, txAmount, energyLimit, energyPrice, payload)
	signer := types.NewNucleusSigner(networkID)
	signedTx, err := types.SignTx(tx, signer, vaultKey)
	if err != nil {
		t.Fatal("can't sign vault funding tx:", err)
	}
	return signedTx
}

// nextNonce generates the nonce of a funding transaction.
func (v *vault) nextNonce() uint64 {
	v.mu.Lock()
	defer v.mu.Unlock()

	nonce := v.nonce
	v.nonce++
	return nonce
}

var (
	predeployedVaultContractSrc = `
pragma solidity ^0.4.6;

// The vault contract is used in the hive rpc-tests suite.
// From this preallocated contract accounts that are created
// during the tests are funded.
contract Vault {
    event Send(address indexed, uint);

    // sendSome send 'amount' wei 'to'
    function sendSome(address to, uint amount) {
        if (to.send(amount)) {
            Send(to, amount);
        }
    }
}`
	// vault ABI
	predeployedVaultABI = `[{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"name":"sendSome","outputs":[],"payable":false,"type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"","type":"address"},{"indexed":false,"name":"","type":"uint256"}],"name":"Send","type":"event"}]`
)
