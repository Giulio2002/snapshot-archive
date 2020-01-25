## Snapshot Archive

This program allows to convert a snapshot at a certain block from a geth database and convert into in turbo-geth format so that it can be reused without downloading the blockchain twice.

## Build

```
make build
```

binaries are generated into the bin/ folder

## test

```
make test
```
for testing
## Command options

```
    --chaindata path to go-ethereum's chaindata
    --out the resultant boltDB output
    --max-operations-per-transaction max operation per tx in boltDB (PUT operations)
    --block-number the number of the block we need to take the snapshot from
```