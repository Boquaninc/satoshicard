package util

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"os/exec"

	"github.com/iden3/go-iden3-crypto/mimc7"
)

const (
	CARDS_NUM = 52
)

type (
	Proof struct {
		A []string   `json:"a"`
		B [][]string `json:"b"`
		C []string   `json:"c"`
	}

	ProofJson struct {
		Scheme string `json:"scheme"`
		Curve  string `json:"curve"`
		P      Proof  `json:"proof"`
	}
)

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
	originCards := make([]int, CARDS_NUM)
	for index := range originCards {
		originCards[index] = index + 1
	}
	seed := new(big.Int)
	seed = seed.Mul(number1, number2)

	gotCards := make([]int, 10)
	for i := 0; i < len(gotCards); i++ {
		seed = seed.Mul(seed, seed)
		gotOriginBigIndex := seed.Mod(seed, big.NewInt(int64(len(originCards)-1-i)))
		gotOriginIndex := int(gotOriginBigIndex.Int64())
		originCards[gotOriginIndex], originCards[len(originCards)-1-i] = originCards[len(originCards)-1-i], originCards[gotOriginIndex]
		gotCards[i] = originCards[len(originCards)-1-i]
	}
	return gotCards
}

func GetProof(number1 *big.Int, number2 *big.Int, winHash *big.Int, factor *big.Int) (*Proof, error) {

	cmd := exec.Command("sh", "scripts/gen_proof.sh", number1.String(), number2.String(), winHash.String(), factor.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	log.Println("Computing witness :", string(output))

	content, err := ioutil.ReadFile("./circuits/proof.json")
	if err != nil {
		return nil, err
	}

	pj := &ProofJson{}

	if err := json.Unmarshal(content, &pj); err != nil {
		return nil, err
	}
	return &pj.P, nil
}

func GetHash(preimage *big.Int) *big.Int {
	return mimc7.MIMC7Hash(preimage, big.NewInt(0))
}
