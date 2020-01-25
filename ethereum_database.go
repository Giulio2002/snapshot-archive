package main

import (
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
)

type EthereumDatabase struct {
	db ethdb.KeyValueStore
}

func NewEthereumDatabase(db ethdb.KeyValueStore) EthereumDatabase {
	return EthereumDatabase{
		db: db,
	}
}

func NewEthereumDatabaseFromChainData(chaindata string, freezer string) (EthereumDatabase, error) {
	lvldb, err := leveldb.New(chaindata, 0, 0, "")
	db, err := rawdb.NewDatabaseWithFreezer(lvldb, freezer, "")
	if err != nil {
		return EthereumDatabase{}, err
	}

	return EthereumDatabase{
		db: db,
	}, nil
}
