package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ledgerwatch/turbo-geth/common"
	"github.com/ledgerwatch/turbo-geth/common/dbutils"
	"github.com/ledgerwatch/turbo-geth/core/rawdb"
	"github.com/ledgerwatch/turbo-geth/core/types/accounts"
	"github.com/ledgerwatch/turbo-geth/ethdb"
	"github.com/ledgerwatch/turbo-geth/trie"
)

// These functions are taken from github.com/ledgerwatch/turbo-geth/cmd/state
func checkRoots(stateDb ethdb.Database, rootHash common.Hash, blockNum uint64) {
	startTime := time.Now()
	var err error
	if blockNum > 0 {
		t := trie.New(rootHash)
		r := trie.NewResolver(0, true, blockNum)
		key := []byte{}
		req := t.NewResolveRequest(nil, key, 0, rootHash[:])
		fmt.Printf("new resolve request for root block with hash %x\n", rootHash)
		r.AddRequest(req)
		if err = r.ResolveWithDb(stateDb, blockNum); err != nil {
			fmt.Printf("%v\n", err)
		}
		fmt.Printf("Trie computation took %v\n", time.Since(startTime))
	} else {
		fmt.Printf("block number is unknown, account trie verification skipped\n")
	}
	startTime = time.Now()
	roots := make(map[common.Hash]*accounts.Account)
	incarnationMap := make(map[uint64]int)
	err = stateDb.Walk(dbutils.StorageBucket, nil, 0, func(k, v []byte) (bool, error) {
		var addrHash common.Hash
		copy(addrHash[:], k[:32])
		if _, ok := roots[addrHash]; !ok {
			if enc, _ := stateDb.Get(dbutils.AccountsBucket, addrHash[:]); enc == nil {
				roots[addrHash] = nil
			} else {
				var account accounts.Account
				if err = account.DecodeForStorage(enc); err != nil {
					return false, err
				}
				roots[addrHash] = &account
				incarnationMap[account.Incarnation]++
			}
		}

		return true, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Incarnation map: %v\n", incarnationMap)
	for addrHash, account := range roots {
		if account != nil {
			st := trie.New(account.Root)
			sr := trie.NewResolver(32, false, blockNum)
			key := []byte{}
			contractPrefix := make([]byte, common.HashLength+common.IncarnationLength)
			copy(contractPrefix, addrHash[:])
			binary.BigEndian.PutUint64(contractPrefix[common.HashLength:], account.Incarnation^^uint64(0))
			streq := st.NewResolveRequest(contractPrefix, key, 0, account.Root[:])
			sr.AddRequest(streq)
			err = sr.ResolveWithDb(stateDb, blockNum)
			if err != nil {
				fmt.Printf("%x: %v\n", addrHash, err)
				fmt.Printf("incarnation: %d, account.Root: %x\n", account.Incarnation, account.Root)
			}
		}
	}
	fmt.Printf("Storage trie computation took %v\n", time.Since(startTime))
}

func VerifySnapshot(blockNum uint64, ethDb *ethdb.BoltDatabase) {
	defer ethDb.Close()
	hash := rawdb.ReadHeadBlockHash(ethDb)
	number := rawdb.ReadHeaderNumber(ethDb, hash)
	var currentBlockNr uint64
	var preRoot common.Hash
	if number != nil {
		header := rawdb.ReadHeader(ethDb, hash, *number)
		currentBlockNr = *number
		preRoot = header.Root
	}
	fmt.Printf("Block number: %d\n", currentBlockNr)
	fmt.Printf("Block root hash: %x\n", preRoot)
	checkRoots(ethDb, preRoot, currentBlockNr)
}
