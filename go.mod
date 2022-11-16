module satoshicard

go 1.19

replace github.com/btcsuite/btcutil => g.mempool.com/3rds/btcutil v0.0.0-20200525032747-a3435748dbe8

replace github.com/btcsuite/btcd => g.mempool.com/3rds/btcd v0.0.0-20210105041900-20dbf124da32

require (
	github.com/btcsuite/btcd v0.0.0-00010101000000-000000000000
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/tyler-smith/go-bip39 v1.1.0
)

require (
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
)
