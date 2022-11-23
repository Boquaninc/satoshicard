package util

import (
	"encoding/hex"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/wire"
	"github.com/sCrypt-Inc/go-bt/v2/sighash"
	"github.com/sCrypt-Inc/go-scryptlib"
)

type HashTimeLockOpenUnlockContext struct {
	Contract *scryptlib.Contract
	Preimage *big.Int
	Key      *btcec.PrivateKey
}

func NewHashTimeLockOpenUnlockContext(path string, preimage *big.Int, Key *btcec.PrivateKey) *HashTimeLockOpenUnlockContext {
	return &HashTimeLockOpenUnlockContext{
		Contract: LoadDesc(path),
		Preimage: preimage,
		Key:      Key,
	}
}

func (ctx *HashTimeLockOpenUnlockContext) SetUnlockScript(msgTx *wire.MsgTx, index int, txPoint *TxInPoint) {
	sig := GetSig(msgTx, index, txPoint.LockScript, uint64(txPoint.Value), txPoint.HashType, ctx.Key)

	scryptlibSig, err := scryptlib.NewSigFromDECBytes(sig, sighash.Flag(txPoint.HashType))
	if err != nil {
		panic(err)
	}
	method := "open"
	err = ctx.Contract.SetPublicFunctionParams(
		method,
		map[string]scryptlib.ScryptType{
			"preimage": scryptlib.NewIntFromBigInt(ctx.Preimage),
			"sig":      scryptlibSig,
		},
	)
	if err != nil {
		panic(err)
	}

	unlockScript, err := ctx.Contract.GetUnlockingScript(method)
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

type HashTimeLockOverTimeContext struct {
	Key      *btcec.PrivateKey
	Contract *scryptlib.Contract
}

func NewHashTimeLockOverTimeContext(path string, Key *btcec.PrivateKey) *HashTimeLockOverTimeContext {
	return &HashTimeLockOverTimeContext{
		Contract: LoadDesc(path),
		Key:      Key,
	}
}

func (ctx *HashTimeLockOverTimeContext) SetUnlockScript(msgTx *wire.MsgTx, index int, txPoint *TxInPoint) {
	sig := GetSig(msgTx, index, txPoint.LockScript, uint64(txPoint.Value), txPoint.HashType, ctx.Key)
	preimage := Bip143PreImage(msgTx, index, txPoint.LockScript, uint64(txPoint.Value), txPoint.HashType)

	scryptlibSig, err := scryptlib.NewSigFromDECBytes(sig, sighash.Flag(txPoint.HashType))
	if err != nil {
		panic(err)
	}
	method := "overtime"
	err = ctx.Contract.SetPublicFunctionParams(
		method,
		map[string]scryptlib.ScryptType{
			"txPreimage": scryptlib.NewSigHashPreimage(preimage),
			"sig":        scryptlibSig,
		},
	)
	if err != nil {
		panic(err)
	}

	unlockScript, err := ctx.Contract.GetUnlockingScript(method)
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
