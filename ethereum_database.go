package main

import "github.com/ethereum/go-ethereum/ethdb"

type EthereumDatabase struct {
	db ethdb.KeyValueStore
}

func NewEthereumDatabase(db ethdb.KeyValueStore) EthereumDatabase {
	return EthereumDatabase{
		db: db,
	}
}
