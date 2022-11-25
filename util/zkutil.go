package util

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"os/exec"
	"strings"

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

var CardCode2Str []string = []string{
	"not a card",
	"Ace of Diamonds",
	"Two of Diamonds",
	"Three of Diamonds",
	"Four of Diamonds",
	"Five of Diamonds",
	"Six of Diamonds",
	"Seven of Diamonds",
	"Eight of Diamonds",
	"Nine of Diamonds",
	"Ten of Diamonds",
	"Jack of Diamonds",
	"Queen of Diamonds",
	"King of Diamonds",
	"Ace of Clubs",
	"Two of Clubs",
	"Three of Clubs",
	"Four of Clubs",
	"Five of Clubs",
	"Six of Clubs",
	"Seven of Clubs",
	"Eight of Clubs",
	"Nine of Clubs",
	"Ten of Clubs",
	"Jack of Clubs",
	"Queen of Clubs",
	"King of Clubs",
	"Ace of Hearts",
	"Two of Hearts",
	"Three of Hearts",
	"Four of Hearts",
	"Five of Hearts",
	"Six of Hearts",
	"Seven of Hearts",
	"Eight of Hearts",
	"Nine of Hearts",
	"Ten of Hearts",
	"Jack of Hearts",
	"Queen of Hearts",
	"King of Hearts",
	"Ace of Spades",
	"Two of Spades",
	"Three of Spades",
	"Four of Spades",
	"Five of Spades",
	"Six of Spades",
	"Seven of Spades",
	"Eight of Spades",
	"Nine of Spades",
	"Ten of Spades",
	"Jack of Spades",
	"Queen of Spades",
	"King of Spades",
}

func GetCardStrs(number1 *big.Int, number2 *big.Int) []string {
	cards := GetCards(number1, number2)
	cardStrs1 := make([]string, 0, 5)
	cardStrs2 := make([]string, 0, 5)
	for i, card := range cards {
		cardStr := CardCode2Str[card]
		if i < 5 {
			cardStrs1 = append(cardStrs1, cardStr)
		} else {
			cardStrs2 = append(cardStrs2, cardStr)
		}
	}
	user1Card := strings.Join(cardStrs1, ",")
	user2Card := strings.Join(cardStrs2, ",")
	return []string{
		user1Card,
		user2Card,
	}
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
		mul := new(big.Int)
		mul = mul.Mul(seed, big.NewInt(2))
		seed = mul
		gotOriginBigIndex := new(big.Int)
		gotOriginBigIndex = gotOriginBigIndex.Mod(mul, big.NewInt(int64(len(originCards)-i)))
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
