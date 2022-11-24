rm -rf satoshicard
go build
rm -rf /work/bin-satoshicard/satoshicard/circuits
rm -rf /work/bin-satoshicard/satoshicard/contract
rm -rf /work/bin-satoshicard/satoshicard/desc
rm -rf /work/bin-satoshicard/satoshicard/satoshicard

cp -r circuits /work/bin-satoshicard/satoshicard/circuits
cp -r contract /work/bin-satoshicard/satoshicard/contract
cp -r desc /work/bin-satoshicard/satoshicard/desc
cp satoshicard /work/bin-satoshicard/satoshicard/satoshicard
