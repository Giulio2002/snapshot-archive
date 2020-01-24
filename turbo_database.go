package main

import "github.com/ledgerwatch/turbo-geth/ethdb"

type TurboDatabase struct {
	db ethdb.BoltDatabase
}

func NewTurboDatabase(db ethdb.BoltDatabase) TurboDatabase {
	return TurboDatabase{
		db: db,
	}
}

func NewTurboDatabaseFromChainData(chaindata string) (TurboDatabase, error) {
	db, err := ethdb.NewBoltDatabase(chaindata)
	if err != nil {
		return TurboDatabase{}, err
	}

	return TurboDatabase{
		db: *db,
	}, nil
}
