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
