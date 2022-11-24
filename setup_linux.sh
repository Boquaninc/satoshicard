rm -rf satoshicard
go build

rm -rf circuits
rm -rf contract
rm -rf desc
cp -r circuits_linux circuits
cp -r contract_linux contract
cp -r desc_linux desc