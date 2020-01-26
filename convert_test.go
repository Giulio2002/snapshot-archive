package main

import (
	"context"
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/ledgerwatch/turbo-geth/ethdb"

	"github.com/Giulio2002/snapshot-archive/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
)

func TestConvert(t *testing.T) {
	var (
		boltDB   = ethdb.NewMemDatabase()
		memDb    = memorydb.New()
		db       = rawdb.NewDatabase(memDb)
		key, _   = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		key1, _  = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
		key2, _  = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
		address  = crypto.PubkeyToAddress(key.PublicKey)
		address1 = crypto.PubkeyToAddress(key1.PublicKey)
		address2 = crypto.PubkeyToAddress(key2.PublicKey)
		theAddr  = common.Address{1}
		gspec    = &core.Genesis{
			Config: &params.ChainConfig{
				HomesteadBlock:      big.NewInt(0),
				EIP150Block:         big.NewInt(0),
				EIP155Block:         big.NewInt(0),
				EIP158Block:         big.NewInt(0),
				ByzantiumBlock:      big.NewInt(0),
				ConstantinopleBlock: big.NewInt(0),
				PetersburgBlock:     big.NewInt(0),
				IstanbulBlock:       big.NewInt(0),
			},
			Alloc: core.GenesisAlloc{
				address:  {Balance: big.NewInt(9000000000000000000)},
				address1: {Balance: big.NewInt(200000000000000000)},
				address2: {Balance: big.NewInt(300000000000000000)},
			},
		}
		signer = types.HomesteadSigner{}
	)
	genesis := gspec.MustCommit(db)
	engine := ethash.NewFaker()
	chainConfig, _, err := core.SetupGenesisBlock(db, gspec)
	if err != nil {
		t.Error(err)
	}
	blockchain, err := core.NewBlockChain(db, &core.CacheConfig{
		TrieDirtyDisabled: true,
	}, chainConfig, engine, vm.Config{}, nil)
	if err != nil {
		t.Error(err)
	}
	_ = blockchain.StateCache().TrieDB()

	contractBackend := backends.NewSimulatedBackend(gspec.Alloc, gspec.GasLimit)
	transactOpts := bind.NewKeyedTransactor(key)
	transactOpts1 := bind.NewKeyedTransactor(key1)
	transactOpts2 := bind.NewKeyedTransactor(key2)

	var tokenContract *contracts.Token

	blocks, _ := core.GenerateChain(gspec.Config, genesis, engine, db, 8, func(i int, block *core.BlockGen) {
		var (
			tx  *types.Transaction
			txs []*types.Transaction
		)

		ctx := context.Background()

		switch i {
		case 0:
			tx, err = types.SignTx(types.NewTransaction(block.TxNonce(address), theAddr, big.NewInt(1000000000000000), 21000, new(big.Int), nil), signer, key)
			err = contractBackend.SendTransaction(ctx, tx)
			if err != nil {
				panic(err)
			}
		case 1:
			tx, err = types.SignTx(types.NewTransaction(block.TxNonce(address), theAddr, big.NewInt(1000000000000000), 21000, new(big.Int), nil), signer, key)
			err = contractBackend.SendTransaction(ctx, tx)
			if err != nil {
				panic(err)
			}
		case 2:
			_, tx, tokenContract, err = contracts.DeployToken(transactOpts, contractBackend, address1)
		case 3:
			tx, err = tokenContract.Mint(transactOpts1, address2, big.NewInt(10))
		case 4:
			tx, err = tokenContract.Transfer(transactOpts2, address, big.NewInt(3))
		case 5:
			// Muliple transactions sending small amounts of ether to various accounts
			var j uint64
			var toAddr common.Address
			nonce := block.TxNonce(address)
			for j = 1; j <= 32; j++ {
				binary.BigEndian.PutUint64(toAddr[:], j)
				tx, err = types.SignTx(types.NewTransaction(nonce, toAddr, big.NewInt(1000000000000000), 21000, new(big.Int), nil), signer, key)
				if err != nil {
					panic(err)
				}
				err = contractBackend.SendTransaction(ctx, tx)
				if err != nil {
					panic(err)
				}
				txs = append(txs, tx)
				nonce++
			}
		case 6:
			_, tx, tokenContract, err = contracts.DeployToken(transactOpts, contractBackend, address1)
			if err != nil {
				panic(err)
			}
			txs = append(txs, tx)
			tx, err = tokenContract.Mint(transactOpts1, address2, big.NewInt(100))
			if err != nil {
				panic(err)
			}
			txs = append(txs, tx)
			// Muliple transactions sending small amounts of ether to various accounts
			var j uint64
			var toAddr common.Address
			for j = 1; j <= 32; j++ {
				binary.BigEndian.PutUint64(toAddr[:], j)
				tx, err = tokenContract.Transfer(transactOpts2, toAddr, big.NewInt(1))
				if err != nil {
					panic(err)
				}
				txs = append(txs, tx)
			}
		case 7:
			var toAddr common.Address
			nonce := block.TxNonce(address)
			binary.BigEndian.PutUint64(toAddr[:], 4)
			tx, err = types.SignTx(types.NewTransaction(nonce, toAddr, big.NewInt(1000000000000000), 21000, new(big.Int), nil), signer, key)
			if err != nil {
				panic(err)
			}
			err = contractBackend.SendTransaction(ctx, tx)
			if err != nil {
				panic(err)
			}
			txs = append(txs, tx)
			binary.BigEndian.PutUint64(toAddr[:], 12)
			tx, err = tokenContract.Transfer(transactOpts2, toAddr, big.NewInt(1))
			if err != nil {
				panic(err)
			}
			txs = append(txs, tx)
		}

		if err != nil {
			panic(err)
		}
		if txs == nil && tx != nil {
			txs = append(txs, tx)
		}

		for _, tx := range txs {
			block.AddTx(tx)
		}
		contractBackend.Commit()
	})

	// BLOCK 1
	if _, err = blockchain.InsertChain(types.Blocks{blocks[0]}); err != nil {
		t.Error(err)
	}

	// BLOCK 2
	if _, err = blockchain.InsertChain(types.Blocks{blocks[1]}); err != nil {
		t.Error(err)
	}

	// BLOCK 3
	if _, err = blockchain.InsertChain(types.Blocks{blocks[2]}); err != nil {
		t.Error(err)
	}

	// BLOCK 4
	if _, err = blockchain.InsertChain(types.Blocks{blocks[3]}); err != nil {
		t.Error(err)
	}
	// BLOCK 5
	if _, err = blockchain.InsertChain(types.Blocks{blocks[4]}); err != nil {
		t.Error(err)
	}
	// BLOCK 6
	if _, err = blockchain.InsertChain(types.Blocks{blocks[5]}); err != nil {
		t.Error(err)
	}
	// BLOCK 7
	if _, err = blockchain.InsertChain(types.Blocks{blocks[6]}); err != nil {
		t.Error(err)
	}
	ethereumDB := NewEthereumDatabase(memDb)
	turboDB := NewTurboDatabase(*boltDB)
	mut := turboDB.db.NewBatch()
	rawDB := rawdb.NewDatabase(ethereumDB.db)
	trieDB := trie.NewDatabase(ethereumDB.db)
	blockTrie, root, _ := newStateTrie(ethereumDB, 7)
	stateDB, _ := state.New(root, state.NewDatabase(rawDB))

	_, _, err = ConvertSnapshot(ethereumDB, turboDB, nil, 1000, trieDB, stateDB, blockTrie, mut)
	if err != nil {
		t.Error(err)
	}

	VerifySnapshot(7, &turboDB.db)
}
