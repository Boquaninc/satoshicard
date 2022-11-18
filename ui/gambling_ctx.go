package ui

import (
	"math/big"
	"satoshicard/util"

	"github.com/sCrypt-Inc/go-scryptlib"
)

type intNumber interface {
	*big.Int | int64 | int | string
}

func NewBigInt[T intNumber](number T) *big.Int {
	var i interface{} = number
	switch t := i.(type) {
	case *big.Int:
		return t
	case int64:
		return big.NewInt(t)
	case string:
		result, ok := big.NewInt(0).SetString(t, 0)
		if !ok {
			panic(t)
		}
		return result
	default:
		panic("not support type")
	}
}

func NewG1PointScryptlibStruct[T1 intNumber, T2 intNumber](x T1, y T2) *scryptlib.Struct {
	return scryptlib.NewStruct(
		"G1Point",
		[]string{"x", "y"},
		map[string]scryptlib.ScryptType{
			"x": scryptlib.NewIntFromBigInt(NewBigInt(x)),
			"y": scryptlib.NewIntFromBigInt(NewBigInt(y))},
		nil,
	)
}

func NewFQ2ScryptlibStruct[T1 intNumber, T2 intNumber](x T1, y T2) *scryptlib.Struct {
	return scryptlib.NewStruct(
		"FQ2",
		[]string{"x", "y"},
		map[string]scryptlib.ScryptType{
			"x": scryptlib.NewIntFromBigInt(NewBigInt(x)),
			"y": scryptlib.NewIntFromBigInt(NewBigInt(y))},
		nil,
	)
}

func NewG2PointScryptlibStruct(x *scryptlib.Struct, y *scryptlib.Struct) *scryptlib.Struct {
	if x.GetTypeString() != "FQ2" || y.GetTypeString() != "FQ2" {
		panic("need both fq2")
	}
	return scryptlib.NewStruct(
		"G2Point",
		[]string{"x", "y"},
		map[string]scryptlib.ScryptType{
			"x": *x,
			"y": *y},
		nil,
	)
}

func NewProofScryptlibStruct(a *scryptlib.Struct, b *scryptlib.Struct, c *scryptlib.Struct) *scryptlib.Struct {
	return scryptlib.NewStruct(
		"Proof",
		[]string{"a", "b", "c"},
		map[string]scryptlib.ScryptType{
			"a": *a,
			"b": *b,
			"c": *c},
		nil,
	)
}

func NewProofScryptlibStructFromProof(proofConfig *util.Proof) *scryptlib.Struct {
	a := NewG1PointScryptlibStruct(proofConfig.A[0], proofConfig.A[1])
	bfq2X := NewFQ2ScryptlibStruct(proofConfig.B[0][0], proofConfig.B[0][1])
	bfq2Y := NewFQ2ScryptlibStruct(proofConfig.B[1][0], proofConfig.B[1][1])
	b := NewG2PointScryptlibStruct(bfq2X, bfq2Y)
	c := NewG1PointScryptlibStruct(proofConfig.C[0], proofConfig.C[1])
	return NewProofScryptlibStruct(a, b, c)
}

type SignGamblingCtx struct {
	Contract *scryptlib.Contract
	Factor   *big.Int
	Number1  *big.Int
	Number2  *big.Int
	WinHash  *big.Int
}

func NewSignGamblingCtx(
	ContractPath string,
	Factor int64,
	Number1 *big.Int,
	Number2 *big.Int,
	WinHash *big.Int,
) *SignGamblingCtx {
	desc, err := scryptlib.LoadDesc(ContractPath)
	if err != nil {
		panic(err)
	}

	contract, err := scryptlib.NewContractFromDesc(desc)
	if err != nil {
		panic(err)
	}
	return &SignGamblingCtx{
		Contract: &contract,
		Factor:   big.NewInt(Factor),
		Number1:  Number1,
		Number2:  Number2,
		WinHash:  WinHash,
	}
}

func (this *SignGamblingCtx) GetProof() (*scryptlib.Struct, error) {
	// fmt.Println("GetProof 1:", this.Number1)
	// fmt.Println("GetProof 2:", this.Number2)
	// fmt.Println("GetProof 3:", this.WinHash)
	// fmt.Println("GetProof 4:", this.Factor)
	// proof, err := util.GetProof(this.Number1, this.Number2, this.WinHash, this.Factor)
	// if err != nil {
	// 	fmt.Println("GetProof -------err:", err)
	// 	return nil, err
	// }
	//     {
	//   "scheme": "g16",
	//   "curve": "bn128",
	//   "proof": {
	//     "a": [
	//       "0x0e8784f99d5b6e507033053f19a123ca9d64813fc75e5bbd8b2af8f935ffa519",
	//       "0x127855b55ac5c07308e34ba89edfb0b165153104faaeff6ec2322e100dc5a9be"
	//     ],
	//     "b": [
	//       [
	//         "0x012e6a9500965a6978bf40118bed2070d3f5c0e3137558bad2aee39d26f3caeb",
	//         "0x04d89437deffaca6afa1c93fc85f28882116f6f393650244a610c0277929d1fa"
	//       ],
	//       [
	//         "0x199daacf33cd725852db62534ad9bca8c7e9c46e9daa7b7845c68a6630f40102",
	//         "0x2afe1ae6f1d9b38fdcee4a3b35432fc6bda637c4616d82d1f7cbad29f9910b34"
	//       ]
	//     ],
	//     "c": [
	//       "0x0fbed969c53a96819d2d2435676187a0a44a3e8d1ec8aab8a0a52982715b2c87",
	//       "0x2c0525f000183cccfe74e1348cb997a49301229726a836bc5fbd6cb286b9ccf5"
	//     ]
	//   },
	//   "inputs": [
	//     "0x226b4640946fa9a4c6fb44e78f3f7c3fb42f5d336f28e31d129994114bae2623",
	//     "0x0000000000000000000000000000000000000000000000000000000000000002"
	//   ]
	// }

	proof := &util.Proof{
		B: make([][]string, 2),
	}
	proof.A = []string{
		"0x0e8784f99d5b6e507033053f19a123ca9d64813fc75e5bbd8b2af8f935ffa519",
		"0x127855b55ac5c07308e34ba89edfb0b165153104faaeff6ec2322e100dc5a9be",
	}
	proof.B[0] = []string{
		"0x012e6a9500965a6978bf40118bed2070d3f5c0e3137558bad2aee39d26f3caeb",
		"0x04d89437deffaca6afa1c93fc85f28882116f6f393650244a610c0277929d1fa",
	}
	proof.B[1] = []string{
		"0x199daacf33cd725852db62534ad9bca8c7e9c46e9daa7b7845c68a6630f40102",
		"0x2afe1ae6f1d9b38fdcee4a3b35432fc6bda637c4616d82d1f7cbad29f9910b34",
	}
	proof.C = []string{
		"0x0fbed969c53a96819d2d2435676187a0a44a3e8d1ec8aab8a0a52982715b2c87",
		"0x2c0525f000183cccfe74e1348cb997a49301229726a836bc5fbd6cb286b9ccf5",
	}
	// util.PrintJson(proof)
	return NewProofScryptlibStructFromProof(proof), nil
}
