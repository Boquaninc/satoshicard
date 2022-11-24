rm -rf satoshicard
go build
./satoshicard \
-help=false \
-listen=0.0.0.0:10001 \
-rpchost=192.168.10.165:19002 \
-rpcusername=regtest \
-rpcpassword=123 \
-gamecontractpath=./desc/satoshicard_release_desc.json \
-lockcontractpath=./desc/satoshicard_timelock_release_desc.json \
-key=ed909bc8d0b35d622a4c3b0c700fce4f1472c533289d5127a782c09c669fb1d7 \
-mode=1