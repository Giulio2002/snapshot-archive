package main

import (
	"flag"
	"fmt"
)

func main() {
	var chaindata = flag.String("chaindata", "", "path to go-ethereum's chaindata")
	var out = flag.String("out", "out", "output path")
	var max = flag.Uint("max-operations-per-transaction", 100000, "the number of operations per transaction in DB")
	var blockNumber = flag.Uint64("block-number", 0, "block number") // replace 0 with latest
	flag.Parse()

	if *chaindata == "" {
		fmt.Println("--chaindata can't be nothing")
		return
	}

	boltDB, err := NewTurboDatabaseFromChainData(*out)
	if err != nil {
		fmt.Println("Cannot initialise bolt database")
		return
	}
	leveldDB, err := NewEthereumDatabaseFromChainData(*chaindata)
	if err != nil {
		fmt.Println("Cannot initialise ethereum database")
		return
	}

	newKey, err := ConvertSnapshot(leveldDB, boltDB, nil, *max, *blockNumber)
	if err != nil {
		fmt.Printf("Convert Operation Failed: %s \n", err.Error())
		return
	}
	for newKey != nil {
		newKey, err = ConvertSnapshot(leveldDB, boltDB, newKey, *max, *blockNumber)
		if err != nil {
			fmt.Printf("Convert Operation Failed: %s \n", err.Error())
			return
		}
	}
	VerifySnapshot(*blockNumber, &boltDB.db)
	fmt.Println("Snapshot converted")
	return
}
