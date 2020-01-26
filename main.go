package main

import (
	"flag"
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
		log.Error("Cannot initialise ethereum database", "err", err.Error())
		return
	}

	boltDB, err := NewTurboDatabaseFromChainData(*out)
	if err != nil {
		log.Error("Cannot initialise bolt database", "err", err.Error())
		return
	}
	var blockNumber uint64
	if *IblockNumber == -1 {
		blockNumber = getBlockNumber(leveldDB)
		log.Info("Latest Block Number", "block", blockNumber)
	} else {
		blockNumber = uint64(*IblockNumber)
	}

	if *chaindata == "" {
		log.Error("--chaindata can't be nothing")
		return
	}

	mut := boltDB.db.NewBatch()
	rawDB := rawdb.NewDatabase(leveldDB.db)
	trieDB := trie.NewDatabase(leveldDB.db)
	t, root, err := newStateTrie(leveldDB, blockNumber)
	if err != nil {
		log.Error("Could not retrieve state trie", "err", err.Error())
		return
	}
	stateDB, err := state.New(root, state.NewDatabase(rawDB))
	if err != nil {
		log.Info("Could not retrieve state trie", "err", err.Error())
		return
	}

	written, start, err := ConvertSnapshot(leveldDB, boltDB, []byte{}, *max, trieDB, stateDB, t, mut)
	if err != nil {
		log.Info("Written", "entries", written)
		log.Error("Convert Operation Failed", "err", err.Error())
		return
	}
	for start != nil {
		wrote, newStart, err := ConvertSnapshot(leveldDB, boltDB, start, *max, trieDB, stateDB, t, mut)
		start = newStart
		written += wrote
		if err != nil {
			log.Info("Written", "entries", written)
			log.Error("Convert Operation Failed", "err", err.Error())
			return
		}
	}
	log.Info("Written", "entries", written)
	log.Info("Snapshot converted")
	return
}
