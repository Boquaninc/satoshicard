./bsv/bin/bitcoin-cli -datadir=./bsv stop
rm -rf ./bsv/regtest
cp -r ./bsv/regtest-bak ./bsv/regtest
nohup ./bsv/bin/bitcoind -datadir=./bsv &
# ./bsv/bin/bitcoin-cli -datadir=./bsv getinfo