package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/trie"
)

func newStateTrie(db EthereumDatabase, blockNumber uint64) (*trie.Trie, common.Hash, error) {
	rawDB := rawdb.NewDatabase(db.db)
	canonical := rawdb.ReadCanonicalHash(rawDB, blockNumber)
	root := rawdb.ReadHeader(rawDB, canonical, blockNumber).Root
	t, err := trie.New(root, trie.NewDatabase(db.db))
	if err != nil {
		return nil, root, err
	}
	return t, root, nil
}
