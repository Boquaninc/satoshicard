rm -rf satoshicard
go build

rm -rf circuits
rm -rf contract
rm -rf desc
cp -r circuits_mac circuits
cp -r contract_mac contract
cp -r desc_mac desc