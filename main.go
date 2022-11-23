package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"satoshicard/conf"
	"satoshicard/server"
	"satoshicard/ui"
	"satoshicard/util"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/sCrypt-Inc/go-scryptlib"
)

type Flags struct {
	Env  string
	Mode int64
}

func WaitInput() {
	waitinput := ""
	fmt.Scanf("%s", &waitinput)
}

func Test1() {
	config := conf.GetConfig()
	uictx := ui.NewUIContext(config, 1)
	hostEvent := &ui.UIEvent{
		Event:  ui.EVENT_HOST,
		Params: "",
	}
	uictx.EventChannel <- hostEvent

	WaitInput()

	preimageEvent := &ui.UIEvent{
		Event:  ui.EVENT_PREIMAGE,
		Params: "22",
	}
	uictx.EventChannel <- preimageEvent

	WaitInput()

	signEvent := &ui.UIEvent{
		Event:  ui.EVENT_SIGN,
		Params: "",
	}
	uictx.EventChannel <- signEvent

	WaitInput()
	publishEvent := &ui.UIEvent{
		Event:  ui.EVENT_PUBLISH,
		Params: "",
	}
	uictx.EventChannel <- publishEvent

	WaitInput()
	openEvent := &ui.UIEvent{
		Event:  ui.EVENT_OPEN,
		Params: "",
	}
	uictx.EventChannel <- openEvent

	WaitInput()
	checkEvent := &ui.UIEvent{
		Event:  ui.EVENT_CHEKC,
		Params: "",
	}
	uictx.EventChannel <- checkEvent

	WaitInput()
	loseEvent := &ui.UIEvent{
		Event:  ui.EVENT_LOSE,
		Params: "",
	}
	uictx.EventChannel <- loseEvent
}

func Test2() {
	config := conf.GetConfig()
	uictx := ui.NewUIContext(config, 2)
	joinEvent := &ui.UIEvent{
		Event:  ui.EVENT_JOIN,
		Params: "127.0.0.1:10001",
	}
	uictx.EventChannel <- joinEvent

	WaitInput()

	preimageEvent := &ui.UIEvent{
		Event:  ui.EVENT_PREIMAGE,
		Params: "27",
	}
	uictx.EventChannel <- preimageEvent

	WaitInput()
	signEvent := &ui.UIEvent{
		Event:  ui.EVENT_SIGN,
		Params: "",
	}
	uictx.EventChannel <- signEvent

	WaitInput()
	openEvent := &ui.UIEvent{
		Event:  ui.EVENT_OPEN,
		Params: "",
	}
	uictx.EventChannel <- openEvent

	// WaitInput()
	// takedepositEvent := &ui.UIEvent{
	// 	Event:  ui.EVENT_TAKEDEPOSIT,
	// 	Params: "",
	// }
	// uictx.EventChannel <- takedepositEvent

	WaitInput()
	checkEvent := &ui.UIEvent{
		Event:  ui.EVENT_CHEKC,
		Params: "",
	}
	uictx.EventChannel <- checkEvent

	WaitInput()
	winEvent := &ui.UIEvent{
		Event:  ui.EVENT_WIN,
		Params: "2",
	}
	uictx.EventChannel <- winEvent
}

func Test3() {
	number1 := big.NewInt(22)
	number2 := big.NewInt(27)
	winhash := util.GetHash(number2)
	factor := big.NewInt(2)

	cards := util.GetCardStrs(number1, number2)
	fmt.Println(cards)

	proof, err := util.GetProof(number1, number2, winhash, factor)
	if err != nil {
		panic(err)
	}
	util.PrintJson(proof)

	proof2, err := util.GetProof(number1, number2, winhash, factor)
	if err != nil {
		panic(err)
	}
	util.PrintJson(proof2)
}

func Test4() {
	config := conf.GetConfig()
	contract := util.LoadDesc(config.LockContractPath)
	number := big.NewInt(27)
	numberHash := util.GetHash(number)

	matureTime := time.Now().Unix() + 60*60
	privateKeyByte, err := hex.DecodeString(config.Key)
	if err != nil {
		panic(err)
	}
	privateKey, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyByte)
	lockConstructorParams1 := map[string]scryptlib.ScryptType{
		"matureTime":   scryptlib.NewInt(matureTime),
		"preimageHash": scryptlib.NewIntFromBigInt(numberHash),
		"pubkey":       scryptlib.NewPubKey(util.ToBecPubkey(pubkey)),
	}
	genesisLockScript := server.GetConstructorLockScript(lockConstructorParams1, contract)
	genesisTxCtx := util.NewTxContext()
	genesisTxCtx.SupplementFeePrivateKey = privateKey
	rpcClient := ui.NewRpcClient(config.RpcClientConfig)
	genesisTxCtx.RpcClient = rpcClient
	genesisAmount := int64(10000)
	genesisTxCtx.AddVout(genesisAmount, genesisLockScript)
	genesisTx := genesisTxCtx.SupplementFeeAndBuildByFaucet()
	txid, err := genesisTxCtx.RpcClient.SendRawTransaction(genesisTx, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Test4 7:", txid.String())

}

func Test5() {
	// pri1, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	// if err != nil {
	// 	panic(err)
	// }
	// pub1 := pri1.PublicKey
	// pri2, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	// if err != nil {
	// 	panic(err)
	// }
	// pub2 := pri2.PublicKey

}

func DoMain() {
	config := conf.GetConfig()
	ui.NewUIContext(config, 0)
}

func main() {
	// uictx := &ui.UIContext{}
	flags := &Flags{}
	flag.StringVar(&flags.Env, "env", "", "")
	flag.Int64Var(&flags.Mode, "mode", 0, "")
	flag.Parse()
	conf.Init(flags.Env)

	switch flags.Mode {
	case 0:
		DoMain()
	case 1:
		Test1()
	case 2:
		Test2()
	case 3:
		Test3()
	case 4:
		Test4()
	default:
		panic("not support mode")
	}
	for {
		time.Sleep(time.Minute)
	}
}
