package main

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ledgerwatch/turbo-geth/common/dbutils"
	"github.com/ledgerwatch/turbo-geth/core/types/accounts"
	"github.com/ledgerwatch/turbo-geth/crypto"
	"github.com/ledgerwatch/turbo-geth/common/debug"
	"github.com/ledgerwatch/turbo-geth/ethdb"
)

var preimagePrefix = []byte("secure-key-")
var emptyCodeHash = crypto.Keccak256Hash(nil)
var emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421").Bytes()

func ConvertSnapshot(from EthereumDatabase, to TurboDatabase, start []byte, maxOperationsPerTransaction uint, blockNumber uint64) ([]byte, error) {
	t, blockchain, err := newStateTrie(from, blockNumber)
	if err != nil {
		return nil, err
	}
	iterator := trie.NewIterator(t.NodeIterator(start))
	var counter uint
	mut := to.db.NewBatch()
	trieDB := trie.NewDatabase(from.db)
	stateDB, err := blockchain.State()
	if err != nil {
		return nil, err
	}

	for iterator.Next() {
		gethAccount := state.Account{} // go-ethereum account
		var tAccount accounts.Account  // turbo-geth account

		if counter == maxOperationsPerTransaction {
			_, err := mut.Commit()
			return iterator.Key, err
		}

		// setup account
		tAccount.Nonce = gethAccount.Nonce
		if gethAccount.Balance != nil {
			tAccount.Balance = *gethAccount.Balance
		}
		// Decode geth account
		err := rlp.Decode(bytes.NewBuffer(iterator.Value), &gethAccount)
		if err != nil {
			return nil, err
		}
		// Storage Bucket
		storageTrie, err := trie.New(gethAccount.Root, trieDB)
		if err != nil {
			return nil, err
		}

		err, isContract := makeStorage(mut, storageTrie, iterator.Key, &counter)

		if err != nil {
			return nil, err
		}
		if isContract {
			tAccount.Root.SetBytes(gethAccount.Root.Bytes())
			tAccount.CodeHash.SetBytes(gethAccount.CodeHash)
			tAccount.Incarnation = 1
			err := makeCode(mut, from, stateDB, iterator.Key)
			counter++
			if err != nil {
				return nil, err
			}
			if debug.IsThinHistory() {
				err := makeContractCode(mut, from, tAccount, stateDB, iterator.Key)
				if err != nil {
					return nil, err
				}
				counter++
			}
		} else {
			tAccount.Incarnation = 0
			tAccount.Root.SetBytes(emptyRoot)
			tAccount.CodeHash = emptyCodeHash
		}
		// Account Bucket
		bytesAccount := make([]byte, tAccount.EncodingLengthForStorage())
		tAccount.EncodeForStorage(bytesAccount)

		err = mut.Put(dbutils.AccountsBucket, iterator.Key, bytesAccount)
		if err != nil {
			return nil, err
		}
		counter++
	}
	_, err = mut.Commit()
	fmt.Println(counter)
	return nil, err
}

func makeStorage(mut ethdb.DbWithPendingMutations, t *trie.Trie, accountKey []byte, counter *uint) (error, bool) {
	iterator := trie.NewIterator(t.NodeIterator(nil))
	var isContract bool

	for iterator.Next() {
		buffer := bytes.NewBuffer(iterator.Value)
		storageValue := []byte{}
		rlp.Decode(buffer, &storageValue)
		err := mut.Put(dbutils.StorageBucket, append(accountKey, iterator.Key...), storageValue)
		if err != nil {
			return err, true
		}
		isContract = true
		*counter++
	}
	return nil, isContract
}

func makeCode(mut ethdb.DbWithPendingMutations, from EthereumDatabase, stateDB *state.StateDB, accountKey []byte) error {
	address, err := getAddress(from, accountKey)
	if err != nil {
		return err
	}
	return mut.Put(dbutils.CodeBucket, accountKey, stateDB.GetCode(address))
}

func makeContractCode(mut ethdb.DbWithPendingMutations, from EthereumDatabase, tAccount accounts.Account, stateDB *state.StateDB, accountKey []byte) error {
	addressHash, err := getAddressHash(from, accountKey)
	if err != nil {
		return err
	}
	return mut.Put(dbutils.ContractCodeBucket, append(addressHash.Bytes(), byte(tAccount.Incarnation)), tAccount.CodeHash.Bytes())
}

func getAddress(from EthereumDatabase, preimage []byte) (common.Address, error) {
	addressBytes, err := from.db.Get(append(preimagePrefix, preimage...))
	return common.BytesToAddress(addressBytes), err
}

func getAddressHash(from EthereumDatabase, preimage []byte) (common.Hash, error) {
	addressBytes, err := from.db.Get(append(preimagePrefix, preimage...))
	return common.BytesToHash(addressBytes), err
}
