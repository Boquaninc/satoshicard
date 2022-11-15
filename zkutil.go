package main

import "math/big"

type Proof struct {
	A []string   `json:"a"`
	B [][]string `json:"b"`
	C []string   `json:"c"`
}

var cards []string = []string{
	"not a card",
	"方片A",
	"方片2",
	"方片3",
	"方片4",
	"方片5",
	"方片6",
	"方片7",
	"方片8",
	"方片9",
	"方片10",
	"方片Jack",
	"方片Queen",
	"方片King",
	"梅花A",
	"梅花2",
	"梅花3",
	"梅花4",
	"梅花5",
	"梅花6",
	"梅花7",
	"梅花8",
	"梅花9",
	"梅花10",
	"梅花Jack",
	"梅花Queen",
	"梅花King",
	"红桃A",
	"红桃2",
	"红桃3",
	"红桃4",
	"红桃5",
	"红桃6",
	"红桃7",
	"红桃8",
	"红桃9",
	"红桃10",
	"红桃Jack",
	"红桃Queen",
	"红桃King",
	"黑桃A",
	"黑桃2",
	"黑桃3",
	"黑桃4",
	"黑桃5",
	"黑桃6",
	"黑桃7",
	"黑桃8",
	"黑桃9",
	"黑桃10",
	"黑桃Jack",
	"黑桃Queen",
	"黑桃King",
}

func GetCards(number1 *big.Int, number2 *big.Int) []int {
	return nil
}

func GetProof(number1 *big.Int, number2 *big.Int, winHash *big.Int, factor *big.Int) (*Proof, error) {
	return nil, nil
}

func GetHash(preimage *big.Int) *big.Int {
	return nil
}
