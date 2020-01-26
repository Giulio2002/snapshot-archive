package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

func main() {
	// Setup logger
	var (
		ostream log.Handler
		glogger *log.GlogHandler
	)

	usecolor := (isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if usecolor {
		output = colorable.NewColorableStderr()
	}
	ostream = log.StreamHandler(output, log.TerminalFormat(usecolor))
	glogger = log.NewGlogHandler(ostream)
	log.Root().SetHandler(glogger)
	glogger.Verbosity(log.LvlInfo)

	// Command
	var chaindata = flag.String("chaindata", "", "path to go-ethereum's chaindata")
	var freezer = flag.String("freezer", "", "path to go-ethereum's freezer")
	var out = flag.String("out", "out", "output path")
	var max = flag.Uint("max-operations-per-transaction", 100000, "the number of operations per transaction in DB")
	var cache = flag.Int("cache", 16, "max cache usage")
	var IblockNumber = flag.Int64("block-number", -1, "block number") // replace 0 with latest

	flag.Parse()
	leveldDB, err := NewEthereumDatabaseFromChainData(*chaindata, *freezer, *cache)
	if err != nil {
		fmt.Printf("Cannot initialise ethereum database: %s\n", err.Error())
		return
	}

	boltDB, err := NewTurboDatabaseFromChainData(*out)
	if err != nil {
		fmt.Println("Cannot initialise bolt database")
		return
	}
	var blockNumber uint64
	if *IblockNumber == -1 {
		blockNumber, err = getBlockNumber(leveldDB)
		if err != nil {
			fmt.Println("Cannot read block number")
			return
		}
		fmt.Printf("Latest Block Number: %d\n", blockNumber)
	} else {
		blockNumber = uint64(*IblockNumber)
	}

	if *chaindata == "" {
		fmt.Println("--chaindata can't be nothing")
		return
	}

	mut := boltDB.db.NewBatch()
	rawDB := rawdb.NewDatabase(leveldDB.db)
	trieDB := trie.NewDatabase(leveldDB.db)
	t, root, err := newStateTrie(leveldDB, blockNumber)
	if err != nil {
		fmt.Printf("Could not retrieve state trie: %s", err.Error())
		return
	}
	iterator := trie.NewIterator(t.NodeIterator(nil))
	stateDB, err := state.New(root, state.NewDatabase(rawDB))
	if err != nil {
		fmt.Printf("Could not retrieve state trie: %s", err.Error())
		return
	}

	written, err := ConvertSnapshot(leveldDB, boltDB, iterator, *max, trieDB, stateDB, t, mut)
	if err != nil {
		fmt.Printf("Written: %d entries\n", written)
		fmt.Printf("Convert Operation Failed: %s \n", err.Error())
		return
	}
	for iterator.Key != nil {
		wrote, err := ConvertSnapshot(leveldDB, boltDB, iterator, *max, trieDB, stateDB, t, mut)
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
