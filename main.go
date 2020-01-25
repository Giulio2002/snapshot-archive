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

	newKey, written, err := ConvertSnapshot(leveldDB, boltDB, []byte{}, *max, *blockNumber)
	if err != nil {
		fmt.Printf("Written: %d entries\n", written)
		fmt.Printf("Convert Operation Failed: %s \n", err.Error())
		return
	}
	for newKey != nil {
		k, wrote, err := ConvertSnapshot(leveldDB, boltDB, newKey, *max, *blockNumber)
		newKey = k
		written += wrote
		if err != nil {
			fmt.Printf("Written: %d entries\n", written)
			fmt.Printf("Convert Operation Failed: %s \n", err.Error())
			return
		}
	}
	fmt.Printf("Written: %d entries\n", written)
	fmt.Println("Snapshot converted")
	return
}
