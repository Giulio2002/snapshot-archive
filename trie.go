package main

import (
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/trie"
)

func newStateTrie(db EthereumDatabase, blockNumber uint64) (*trie.Trie, *core.BlockChain, error) {
	rawDb := rawdb.NewDatabase(db.db)
	blockchain, err := core.NewBlockChain(rawDb, nil, nil, ethash.NewFaker(), vm.Config{}, nil)
	if err != nil {
		return nil, nil, err
	}
	block := blockchain.GetBlockByNumber(uint64(blockNumber))
	t, err := trie.New(block.Header().Root, trie.NewDatabase(rawDb))
	if err != nil {
		return nil, nil, err
	}
	return t, blockchain, nil
}
