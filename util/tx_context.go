package util

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

type TxInPoint struct {
	PreTxid    string
	PreIndex   int64
	Value      int64
	LockScript []byte
	HashType   txscript.SigHashType
}

type UnlockContext interface {
	SetUnlockScript(msgTx *wire.MsgTx, index int, txPoint *TxInPoint)
}

type P2pk2UnlockContext struct {
	Key *btcec.PrivateKey
}

func NewP2pk2UnlockContext(Key *btcec.PrivateKey) *P2pk2UnlockContext {
	return &P2pk2UnlockContext{
		Key: Key,
	}
}

func (ctx *P2pk2UnlockContext) SetUnlockScript(msgTx *wire.MsgTx, index int, txPoint *TxInPoint) {
	sig := GetSig(msgTx, index, txPoint.LockScript, uint64(txPoint.Value), txPoint.HashType, ctx.Key)
	builder := txscript.NewScriptBuilder()
	pubbyte := ctx.Key.PubKey().SerializeCompressed()
	b, err := builder.AddData(sig).AddData(pubbyte).Script()
	if err != nil {
		panic(err)
	}
	msgTx.TxIn[index].SignatureScript = b
}

type VinContext struct {
	TxInPoint     *TxInPoint
	UnlockContext UnlockContext
}

func NewVinContext(txInPoint *TxInPoint, unlockContext UnlockContext) *VinContext {
	return &VinContext{
		TxInPoint:     txInPoint,
		UnlockContext: unlockContext,
	}
}

type TxContext struct {
	LockTime                int64
	Vins                    []*VinContext
	Vouts                   []*wire.TxOut
	RpcClient               *rpcclient.Client
	SupplementFeePrivateKey *btcec.PrivateKey
}

func NewTxContext() *TxContext {
	return &TxContext{
		Vins:  make([]*VinContext, 0, 1),
		Vouts: make([]*wire.TxOut, 0, 1),
	}
}

func (ctx *TxContext) AddVin(txInPoint *TxInPoint, unlockVinContext UnlockContext) {
	vinCtx := NewVinContext(txInPoint, unlockVinContext)
	ctx.Vins = append(ctx.Vins, vinCtx)
}

func (ctx *TxContext) AddVout(amount int64, pkScript []byte) {
	out := wire.NewTxOut(amount, pkScript)
	ctx.Vouts = append(ctx.Vouts, out)
}

func (ctx *TxContext) SupplementFeeAndBuildByFaucet() *wire.MsgTx {
	totalVout := int64(0)
	for _, vout := range ctx.Vouts {
		totalVout += vout.Value
	}
	totalVin := int64(0)
	for _, vin := range ctx.Vins {
		totalVin += vin.TxInPoint.Value
	}
	msgTx := ctx.Build()
	size := msgTx.SerializeSize()
	lack := totalVout + int64(size) - totalVin
	if lack < 0 {
		return msgTx
	}
	if lack < 546 {
		lack = 546
	}
	lack = lack + 180
	client := ctx.RpcClient
	faucetAddr := PrivateKey2Address(ctx.SupplementFeePrivateKey)

	txid, err := client.SendToAddress(faucetAddr, btcutil.Amount(lack))
	if err != nil {
		panic(err)
	}

	tx, err := client.GetRawTransaction(txid)
	if err != nil {
		panic(err)
	}
	faucetMsgTx := tx.MsgTx()
	utxoIndex := int64(-1)
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
		addr, err := PkScript.Address(GetNet())
		if err != nil {
			panic(err)
		}
		if addr.String() != faucetAddrStr {
			continue
		}
		utxoIndex = int64(index)
		suppleFeeUtxo = vout
		break
	}
	if utxoIndex == -1 {
		panic("BuildTxContext SupplementFeeAndSign utxoIndex == -1")
	}
	txPoint := &TxInPoint{
		PreTxid:    txid.String(),
		PreIndex:   utxoIndex,
		Value:      suppleFeeUtxo.Value,
		LockScript: suppleFeeUtxo.PkScript,
		HashType:   txscript.SigHashAll | SigHashForkID,
	}
	p2pk2UnlockContext := NewP2pk2UnlockContext(ctx.SupplementFeePrivateKey)
	ctx.AddVin(txPoint, p2pk2UnlockContext)
	return ctx.Build()
}

func (ctx *TxContext) Build() *wire.MsgTx {
	msgTx := wire.NewMsgTx(2)
	msgTx.TxOut = ctx.Vouts
	for _, vin := range ctx.Vins {
		prehash, err := chainhash.NewHashFromStr(vin.TxInPoint.PreTxid)
		if err != nil {
			panic(err)
		}
		preOutPoint := wire.NewOutPoint(prehash, uint32(vin.TxInPoint.PreIndex))
		vin := wire.NewTxIn(preOutPoint, nil, nil)
		msgTx.AddTxIn(vin)
	}
	msgTx.LockTime = uint32(ctx.LockTime)
	for index, vin := range ctx.Vins {
		vin.UnlockContext.SetUnlockScript(msgTx, index, vin.TxInPoint)
	}
	return msgTx
}
