package ui

import (
	"encoding/hex"
	"satoshicard/util"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/sCrypt-Inc/go-scryptlib"
)

const (
	TXPOINT_VIN_VALUE     = -1
	TXPOINT_VOUT_PRETXID  = ""
	TXPOINT_VOUT_PREINDEX = -1
	TXPOINT_VIN_INDEX     = -1
)

type TxPointType int

const (
	TXPOINT_TYPE_VIN      TxPointType = -1
	TXPOINT_TYPE_P2PKH    TxPointType = 0
	TXPOINT_TYPE_GAMBLING TxPointType = 1
)

type TxOutPoint struct {
	Txid    string
	Index   int
	Value   int64
	Type    TxPointType
	Script  []byte `json:"-"`
	Address string
}

type BuildTxContext struct {
	Vins                    []*TxOutPoint
	Vouts                   []*wire.TxOut
	KeyDb                   map[string]*btcec.PrivateKey
	SupplementFeePrivateKey *btcec.PrivateKey
	RpcClient               *rpcclient.Client
	CtxSet                  map[TxPointType]interface{}
}

func NewBuildTxContext(
	RpcClient *rpcclient.Client,
	SupplementFeePrivateKey *btcec.PrivateKey,
	OtherKeys ...*btcec.PrivateKey,
) *BuildTxContext {
	KeyDb := make(map[string]*btcec.PrivateKey)
	for _, otherKey := range OtherKeys {
		addr := util.PrivateKey2Address(otherKey)
		KeyDb[addr.String()] = otherKey
	}

	addr := util.PrivateKey2Address(SupplementFeePrivateKey)
	KeyDb[addr.String()] = SupplementFeePrivateKey
	return &BuildTxContext{
		Vins:                    make([]*TxOutPoint, 0, 2),
		Vouts:                   make([]*wire.TxOut, 0, 2),
		KeyDb:                   KeyDb,
		SupplementFeePrivateKey: SupplementFeePrivateKey,
		RpcClient:               RpcClient,
		CtxSet:                  make(map[TxPointType]interface{}),
	}
}

func (this *BuildTxContext) AddVin(Vin *TxOutPoint) {
	this.Vins = append(this.Vins, Vin)
}

func (this *BuildTxContext) AddVout(amount int64, pkScirpt []byte) {
	txout := wire.NewTxOut(amount, pkScirpt)
	this.Vouts = append(this.Vouts, txout)
}

func (this *BuildTxContext) SignP2PKH(msgTx *wire.MsgTx, index int, txPoint *TxOutPoint) {
	key := this.KeyDb[txPoint.Address]

	sig := util.GetSig(msgTx, index, txPoint.Script, uint64(txPoint.Value), txscript.SigHashAll|util.SigHashForkID, key)

	builder := txscript.NewScriptBuilder()

	pubbyte := key.PubKey().SerializeCompressed()

	b, err := builder.AddData(sig).AddData(pubbyte).Script()
	if err != nil {
		panic(err)
	}
	msgTx.TxIn[index].SignatureScript = b
}

func (this *BuildTxContext) SignGambling(msgTx *wire.MsgTx, index int, txPoint *TxOutPoint) {
	ctxInterface, ok := this.CtxSet[TXPOINT_TYPE_GAMBLING]
	if !ok {
		panic("ctx not found")
	}
	preImage := util.Bip143PreImage(msgTx, index, txPoint.Script, uint64(txPoint.Value), txscript.SigHashAll|util.SigHashForkID)

	signGamblingCtx := ctxInterface.(*SignGamblingCtx)

	proof, err := signGamblingCtx.GetProof()
	if err != nil {
		panic(err)
	}

	method := "run"
	signGamblingCtx.Contract.SetPublicFunctionParams(
		method,
		map[string]scryptlib.ScryptType{
			"txPreimage": scryptlib.NewSigHashPreimage(preImage),
			"proof":      *proof,
			"number1":    scryptlib.NewIntFromBigInt(signGamblingCtx.Number1),
			"number2":    scryptlib.NewIntFromBigInt(signGamblingCtx.Number2),
			"winHash":    scryptlib.NewIntFromBigInt(signGamblingCtx.WinHash),
			"factor":     scryptlib.NewIntFromBigInt(signGamblingCtx.Factor),
		})

	unlockScript, err := signGamblingCtx.Contract.GetUnlockingScript(method)
	if err != nil {
		panic(err)
	}
	unlockScriptHex := unlockScript.String()
	unlockScriptHexByte, err := hex.DecodeString(unlockScriptHex)
	if err != nil {
		panic(err)
	}
	msgTx.TxIn[index].SignatureScript = unlockScriptHexByte
}

func (this *BuildTxContext) Sign() (*wire.MsgTx, error) {
	msgTx := wire.NewMsgTx(2)

	msgTx.TxOut = this.Vouts
	for _, utxo := range this.Vins {
		prehash, err := chainhash.NewHashFromStr(utxo.Txid)
		if err != nil {
			panic(err)
		}
		preOutPoint := wire.NewOutPoint(prehash, uint32(utxo.Index))
		// fmt.Println("sign preOutPoint:", preOutPoint.String())
		vin := wire.NewTxIn(preOutPoint, nil, nil)
		msgTx.AddTxIn(vin)
	}

	for index, utxo := range this.Vins {
		if utxo.Type != TXPOINT_TYPE_P2PKH {
			continue
		}
		this.SignP2PKH(msgTx, index, utxo)
	}

	for index, utxo := range this.Vins {
		if utxo.Type != TXPOINT_TYPE_GAMBLING {
			continue
		}
		this.SignGambling(msgTx, index, utxo)
	}

	return msgTx, nil
}

func (this *BuildTxContext) SupplementFeeAndSign() (*wire.MsgTx, error) {
	totalVout := int64(0)
	for _, vout := range this.Vouts {
		totalVout += vout.Value
	}
	totalVin := int64(0)
	for _, vin := range this.Vins {
		totalVin += vin.Value
	}
	msgTx, err := this.Sign()
	if err != nil {
		return nil, err
	}
	size := msgTx.SerializeSize()
	lack := totalVout + int64(size) - totalVin
	if lack < 0 {
		// fmt.Println("BuildTxContext SupplementFeeAndSign 1")
		return msgTx, nil
	}
	if lack < 546 {
		lack = 546
	}
	lack = lack + 180
	client := this.RpcClient

	faucetAddr := util.PrivateKey2Address(this.SupplementFeePrivateKey)

	txid, err := client.SendToAddress(faucetAddr, btcutil.Amount(lack))
	if err != nil {
		panic(err)
	}

	tx, err := client.GetRawTransaction(txid)
	if err != nil {
		panic(err)
	}
	faucetMsgTx := tx.MsgTx()

	utxoIndex := -1
	var suppleFeeUtxo *wire.TxOut = nil
	faucetAddrStr := faucetAddr.String()
	for index, vout := range faucetMsgTx.TxOut {
		if vout.Value != lack {
			continue
		}
		PkScript, err := txscript.ParsePkScript(vout.PkScript)
		if err != nil {
			panic(err)
		}
		addr, err := PkScript.Address(util.GetNet())
		if err != nil {
			panic(err)
		}
		if addr.String() != faucetAddrStr {
			continue
		}
		utxoIndex = index
		suppleFeeUtxo = vout
		break
	}
	if utxoIndex == -1 {
		panic("BuildTxContext SupplementFeeAndSign utxoIndex == -1")
	}

	txPoint := &TxOutPoint{
		Txid:    txid.String(),
		Index:   utxoIndex,
		Value:   suppleFeeUtxo.Value,
		Type:    TXPOINT_TYPE_P2PKH,
		Script:  suppleFeeUtxo.PkScript,
		Address: faucetAddrStr,
	}
	this.AddVin(txPoint)
	return this.Sign()
}
