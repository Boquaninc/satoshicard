package util

import (
	"encoding/hex"
	"math/big"

	"github.com/btcsuite/btcd/wire"
	"github.com/sCrypt-Inc/go-scryptlib"
)

type HashTimeLockOpenUnlockContext struct {
	Contract *scryptlib.Contract
	Preimage *big.Int
	// Key      *btcec.PrivateKey
}

func NewHashTimeLockOpenUnlockContext(path string, preimage *big.Int) *HashTimeLockOpenUnlockContext {
	return &HashTimeLockOpenUnlockContext{
		Contract: LoadDesc(path),
		Preimage: preimage,
	}
}

func (ctx *HashTimeLockOpenUnlockContext) SetUnlockScript(msgTx *wire.MsgTx, index int, txPoint *TxInPoint) {
	method := "open"
	err := ctx.Contract.SetPublicFunctionParams(
		method,
		map[string]scryptlib.ScryptType{
			"preimage": scryptlib.NewIntFromBigInt(ctx.Preimage),
			// "sig":      scryptlibSig,
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
