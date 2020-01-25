package main

import (
	"github.com/ethereum/go-ethereum/common/fdlimit"
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

func NewEthereumDatabaseFromChainData(chaindata string, freezer string) (EthereumDatabase, error) {
	handles, err := makeDatabaseHandles()
	if err != nil {
		return EthereumDatabase{}, err
	}
	db, err := rawdb.NewLevelDBDatabaseWithFreezer(chaindata, 0, handles, freezer, "")
	if err != nil {
		return EthereumDatabase{}, err
	}

	return EthereumDatabase{
		db: db,
	}, nil
}

func makeDatabaseHandles() (int, error) {
	limit, err := fdlimit.Maximum()
	if err != nil {
		return 0, err
	}
	raised, err := fdlimit.Raise(uint64(limit))
	if err != nil {
		return 0, err
	}
	return int(raised / 2), nil // Leave half for networking and other stuff
}
