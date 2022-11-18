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
	proof, err := util.GetProof(this.Number1, this.Number2, this.WinHash, this.Factor)
	if err != nil {
		return nil, err
	}
	proof.A = []string{
		"0x2dddc80b67d41ae5d8a52dc402f08cdeb372f8e3cad434f7e2f33dd39bcdfb0c",
		"0x25643783288f7794c4f814abf086d76507f1783d1e793900ce7eb78a6ce44659",
	}
	proof.B[0] = []string{
		"0x05466a9d214efa30a0b0976cbf01209f6ef031c95eb7f17a2a6a54c82b1df215",
		"0x15e3b5fe63bb130dfa97b88733ce45e7c41bb33dc189cf8f232466bb222d4fec",
	}
	proof.B[1] = []string{
		"0x20e7a7610f15a639b4eb9704922e47fd3b1279b39518f2da52e9843a7aba59eb",
		"0x3003465623d8f9c38952b4abf4c2fefddb5449f12f9619927943c43ecc1bedec",
	}
	proof.C = []string{
		"0x01cae695071bbeb0ca0adfe716a9948620e3faabaa6aa7521654d9ab253b2a51",
		"0x17f04306cf0d093cf2d544d19aa1aea9f52df35dc7521e773115eb489960997a",
	}
	// util.PrintJson(proof)
	return NewProofScryptlibStructFromProof(proof), nil
}
