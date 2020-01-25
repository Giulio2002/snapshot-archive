package main

import (
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
)

type EthereumDatabase struct {
	db ethdb.KeyValueStore
}

func NewEthereumDatabase(db ethdb.KeyValueStore) EthereumDatabase {
	return EthereumDatabase{
		db: db,
	}
}

func NewEthereumDatabaseFromChainData(chaindata string) (EthereumDatabase, error) {
	db, err := rawdb.NewLevelDBDatabaseWithFreezer(chaindata, 0, 0, "", "")
	if err != nil {
		return EthereumDatabase{}, err
	}

	return EthereumDatabase{
		db: db,
	}, nil
}
