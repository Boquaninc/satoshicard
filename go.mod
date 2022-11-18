module satoshicard

go 1.19

replace github.com/btcsuite/btcutil => g.mempool.com/3rds/btcutil v0.0.0-20200525032747-a3435748dbe8

replace github.com/btcsuite/btcd => g.mempool.com/3rds/btcd v0.0.0-20210105041900-20dbf124da32

replace github.com/sCrypt-Inc/go-scryptlib => ../go-scryptlib

require (
	github.com/btcsuite/btcd v0.0.0-00010101000000-000000000000
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/iden3/go-iden3-crypto v0.0.13
	github.com/libsv/go-bk v0.1.6
	github.com/sCrypt-Inc/go-scryptlib v0.0.0-00010101000000-000000000000
	github.com/tyler-smith/go-bip39 v1.1.0
)

require (
	github.com/aead/siphash v1.0.1 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/btcsuite/go-socks v0.0.0-20170105172521-4720035b7bfd // indirect
	github.com/btcsuite/goleveldb v0.0.0-20160330041536-7834afc9e8cd // indirect
	github.com/btcsuite/snappy-go v0.0.0-20151229074030-0bdef8d06723 // indirect
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792 // indirect
	github.com/btcsuite/winsvc v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jessevdk/go-flags v0.0.0-20141203071132-1679536dcc89 // indirect
	github.com/jrick/logrotate v1.0.0 // indirect
	github.com/kkdai/bstream v0.0.0-20161212061736-f391b8402d23 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sCrypt-Inc/go-bt/v2 v2.1.0-beta.8 // indirect
	github.com/thoas/go-funk v0.9.2 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
)
