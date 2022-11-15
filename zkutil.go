package main

import "math/big"

type Proof struct {
	A []string   `json:"a"`
	B [][]string `json:"b"`
	C []string   `json:"c"`
}

func GetProof(path string) (*Proof, error) {
	return nil, nil
}

func GetHash(path string, preimage *big.Int) *big.Int {
	return nil
}
