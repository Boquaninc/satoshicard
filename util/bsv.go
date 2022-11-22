package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/libsv/go-bk/bec"
	"github.com/sCrypt-Inc/go-scryptlib"
	"github.com/tyler-smith/go-bip39"
)

const (
	BADGE_FLAG    = "badge"
	sigHashMask   = 0x1f
	SigHashForkID = 0x40

	OP_MAX_SINGLE_BYTE_PUSH_DATA = byte(0x4b)

	OP_PUSH_DATA_1 = byte(0x4c)

	OP_PUSH_DATA_2 = byte(0x4d)

	OP_PUSH_DATA_4 = byte(0x4e)

	OP_PUSH_DATA_1_MAX = uint64(255)

	OP_PUSH_DATA_2_MAX = uint64(65535)

	BADGE_CODE_PART_HEX_PREFIX = "5101400100015101b101b26114"
	BADGE_CODE_PART_HEX_SUFFIX = "005179517a7561587905626164676587695979a9517987695a795a79ac7777777777777777777777"
	PUBKEY_HASH_LEN            = 20
	HEX_PUBKEY_HASH_LEN        = 2 * PUBKEY_HASH_LEN
	BADGE_CODE_PART_LEN        = len(BADGE_CODE_PART_HEX_PREFIX)/2 + len(BADGE_CODE_PART_HEX_SUFFIX)/2 + PUBKEY_HASH_LEN
	BADGE_DATA_PART_HEX_PRIFIX = "6a08"
	BADGE_DATA_LEN             = 8
	BADGE_DATA_PART_LEN        = len(BADGE_DATA_PART_HEX_PRIFIX)/2 + BADGE_DATA_LEN // op_return + op_8 + BADGE_DATA_LEN
	BADGE_LOCKING_SCRIPT_LEN   = BADGE_CODE_PART_LEN + BADGE_DATA_PART_LEN

	TX_VERSION = 2

	NFT_VIN_PARSED_OPCODES_LEN = 11
	NFT_VOUT_LEN               = 2188
)

var gNet *chaincfg.Params = &chaincfg.RegressionNetParams

func GetNet() *chaincfg.Params {
	return gNet
}

func SeserializeMsgTx(msgtx *wire.MsgTx) string {
	buf := make([]byte, 0, msgtx.SerializeSize())
	buff := bytes.NewBuffer(buf)
	msgtx.Serialize(buff)
	rawtxByte := buff.Bytes()
	rawtx := hex.EncodeToString(rawtxByte)
	return rawtx
}

func DeserializeRawTx(rawtx string) *wire.MsgTx {
	serializedTx, err := hex.DecodeString(rawtx)
	if err != nil {
		panic(err)
	}
	msgtx := wire.NewMsgTx(TX_VERSION)
	err = msgtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		panic(err)
	}
	return msgtx
}

func WritePushDataScript(buf *bytes.Buffer, data []byte) error {
	size := len(data)
	var err error
	if size <= int(OP_MAX_SINGLE_BYTE_PUSH_DATA) {
		_, err = buf.Write([]byte{byte(size)}) // Single byte push
	} else if size < int(OP_PUSH_DATA_1_MAX) {
		_, err = buf.Write([]byte{OP_PUSH_DATA_1, byte(size)})
	} else if size < int(OP_PUSH_DATA_2_MAX) {
		_, err = buf.Write([]byte{OP_PUSH_DATA_2})
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, uint16(size))
	} else {
		_, err = buf.Write([]byte{OP_PUSH_DATA_4})
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, uint32(size))
	}
	if err != nil {
		return err
	}

	_, err = buf.Write(data)
	return err
}

func WriteVarInt(w io.Writer, pver uint32, val uint64) error {
	if val < 0xfd {
		return binary.Write(w, binary.LittleEndian, uint8(val))
	}

	if val <= math.MaxUint16 {
		err := binary.Write(w, binary.LittleEndian, uint8(0xfd))
		if err != nil {
			return err
		}
		return binary.Write(w, binary.LittleEndian, uint16(val))
	}

	if val <= math.MaxUint32 {
		err := binary.Write(w, binary.LittleEndian, uint8(0xfe))
		if err != nil {
			return err
		}
		return binary.Write(w, binary.LittleEndian, uint32(val))
	}

	err := binary.Write(w, binary.LittleEndian, uint8(0xff))
	if err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, val)
}

func WriteVarBytes(w io.Writer, pver uint32, bytes []byte) error {
	slen := uint64(len(bytes))
	err := WriteVarInt(w, pver, slen)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

func Sha256(b []byte) []byte {
	hasher := sha256.New()
	_, err := hasher.Write(b)
	if err != nil {
		panic(err)
	}
	return hasher.Sum(nil)
}

func Bip143SinglePreVout(in *wire.TxIn) []byte {
	var buf bytes.Buffer
	_, err := buf.Write(in.PreviousOutPoint.Hash.CloneBytes())
	if err != nil {
		panic(err)
	}
	err = binary.Write(&buf, binary.LittleEndian, in.PreviousOutPoint.Index)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func Bip143PreVoutHash(tx *wire.MsgTx, hashType txscript.SigHashType) []byte {
	if hashType&txscript.SigHashAnyOneCanPay != 0 {
		return make([]byte, 32)
	}
	var buf bytes.Buffer
	for _, in := range tx.TxIn {
		singlePreVoutHash := Bip143SinglePreVout(in)
		_, err := buf.Write(singlePreVoutHash)
		if err != nil {
			panic(err)
		}
	}
	return Sha256(Sha256(buf.Bytes()))
}

func Bip143SequenceHash(tx *wire.MsgTx, hashType txscript.SigHashType) []byte {
	if hashType&txscript.SigHashAnyOneCanPay != 0 ||
		hashType&sigHashMask == txscript.SigHashSingle ||
		hashType&sigHashMask == txscript.SigHashNone {
		return make([]byte, 32)
	}
	var buf bytes.Buffer
	for _, in := range tx.TxIn {
		err := binary.Write(&buf, binary.LittleEndian, in.Sequence)
		if err != nil {
			panic(err)
		}
	}
	return Sha256(Sha256(buf.Bytes()))
}

func Bip143VoutHash(tx *wire.MsgTx, hashType txscript.SigHashType, index int) []byte {
	if hashType&txscript.SigHashSingle != txscript.SigHashSingle && hashType&txscript.SigHashNone != txscript.SigHashNone {
		var buf bytes.Buffer
		for _, out := range tx.TxOut {
			err := binary.Write(&buf, binary.LittleEndian, uint64(out.Value))
			if err != nil {
				panic(err)
			}
			err = WriteVarBytes(&buf, 0, out.PkScript)
			if err != nil {
				panic(err)
			}
		}
		return Sha256(Sha256(buf.Bytes()))
	}
	if hashType&sigHashMask == txscript.SigHashSingle && index < len(tx.TxOut) {
		var buf bytes.Buffer
		out := tx.TxOut[index]
		err := binary.Write(&buf, binary.LittleEndian, uint64(out.Value))
		if err != nil {
			panic(err)
		}
		err = WriteVarBytes(&buf, 0, out.PkScript)
		if err != nil {
			panic(err)
		}
		return Sha256(Sha256(buf.Bytes()))
	}
	return make([]byte, 32)
}

func Bip143PreImage(tx *wire.MsgTx, index int, lockScript []byte, value uint64, hashType txscript.SigHashType) []byte {
	if index > len(tx.TxIn)-1 {
		errStr := fmt.Sprintf("signatureHash error: index %d but %d txins", index, len(tx.TxIn))
		panic(errStr)
	}
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, tx.Version)
	if err != nil {
		panic(err)
	}
	preVoutHash := Bip143PreVoutHash(tx, hashType)
	_, err = buf.Write(preVoutHash)
	if err != nil {
		panic(err)
	}
	sequenceHash := Bip143SequenceHash(tx, hashType)
	_, err = buf.Write(sequenceHash)
	if err != nil {
		panic(err)
	}
	singlePreVoutHash := Bip143SinglePreVout(tx.TxIn[index])
	_, err = buf.Write(singlePreVoutHash)
	if err != nil {
		panic(err)
	}
	err = WriteVarBytes(&buf, 0, lockScript)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&buf, binary.LittleEndian, value)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&buf, binary.LittleEndian, tx.TxIn[index].Sequence)
	if err != nil {
		panic(err)
	}
	voutHash := Bip143VoutHash(tx, hashType, index)
	_, err = buf.Write(voutHash)
	if err != nil {
		panic(err)
	}

	binary.Write(&buf, binary.LittleEndian, tx.LockTime)
	binary.Write(&buf, binary.LittleEndian, uint32(hashType|SigHashForkID))
	return buf.Bytes()
}

func Bip143Hash(tx *wire.MsgTx, index int, lockScript []byte, value uint64, hashType txscript.SigHashType) []byte {
	return Sha256(Sha256(Bip143PreImage(tx, index, lockScript, value, hashType)))
}

func SignRFC6979(privateKey *btcec.PrivateKey, hash []byte, hashType txscript.SigHashType) []byte {
	sigNature, err := privateKey.Sign(hash)
	if err != nil {
		panic(err)
	}
	return append(sigNature.Serialize(), byte(hashType))
}

func GetSig(
	tx *wire.MsgTx,
	index int,
	lockScript []byte,
	value uint64,
	hashType txscript.SigHashType,
	privateKey *btcec.PrivateKey,
) []byte {
	hash := Bip143Hash(tx, index, lockScript, value, hashType)
	return SignRFC6979(privateKey, hash, hashType)
}

func ParsePath(path string) []uint32 {
	if len(path) < 3 {
		panic("ParsePath err 1")
	}
	if !strings.HasPrefix(path, "m/") {
		panic("path dose not start with m/")
	}
	ss := strings.Split(path[2:], "/")
	if len(ss) != 3 {
		panic("path should be three step")
	}
	result := make([]uint32, 0, 3)
	for _, s := range ss {
		step := int64(-1)
		if strings.HasSuffix(s, "'") {
			if len(s) < 2 {
				panic("path not num")
			}
			stepTmp, err := strconv.ParseInt(s[:len(s)-1], 10, 64)
			if err != nil {
				panic("path not num")
			}
			step = stepTmp + hdkeychain.HardenedKeyStart
		} else {
			stepTmp, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				panic("path not num")
			}
			step = stepTmp
		}
		result = append(result, uint32(step))
	}
	return result
}

func GetKeyByPath(extendKey *hdkeychain.ExtendedKey, path []uint32) *hdkeychain.ExtendedKey {
	extendKeyLocal := extendKey
	for _, step := range path {
		extendKeyTmp, err := extendKeyLocal.Child(step)
		if err != nil {
			panic(err)
		}
		extendKeyLocal = extendKeyTmp
	}
	return extendKeyLocal
}

func MnemonicWord2Key(mnemonicWords string, pathStr string) *hdkeychain.ExtendedKey {
	seed := bip39.NewSeed(mnemonicWords, "")
	extendKey, err := hdkeychain.NewMaster(seed, &chaincfg.RegressionNetParams)
	if err != nil {
		panic(err)
	}
	path := ParsePath(pathStr)
	return GetKeyByPath(extendKey, path)
}

func PrivateKey2Address(key *btcec.PrivateKey) btcutil.Address {
	return Pubkey2Address(key.PubKey())
}

func Pubkey2Address(key *btcec.PublicKey) btcutil.Address {
	pkHash := btcutil.Hash160(key.SerializeCompressed())
	addr, err := btcutil.NewAddressPubKeyHash(pkHash, gNet)
	if err != nil {
		panic(err)
	}
	return addr
}

func ParseBadgeVoutScript(script []byte) (btcutil.Address, int64) {
	if len(script) != BADGE_LOCKING_SCRIPT_LEN {
		panic("ParseBadgeVoutScript 1")
	}

	hexScript := hex.EncodeToString(script)
	if !strings.HasPrefix(hexScript, BADGE_CODE_PART_HEX_PREFIX) {
		panic("ParseBadgeVoutScript 2")
	}

	index := strings.Index(hexScript, BADGE_CODE_PART_HEX_SUFFIX)
	if index != len(BADGE_CODE_PART_HEX_PREFIX)+HEX_PUBKEY_HASH_LEN {
		panic("ParseBadgeVoutScript 3")
	}
	if hexScript[146] != 0x36 ||
		hexScript[147] != 0x61 ||
		hexScript[148] != 0x30 ||
		hexScript[149] != 0x38 {
		panic("ParseBadgeVoutScript 4")
	}
	address, err := btcutil.NewAddressPubKeyHash(script[len(BADGE_CODE_PART_HEX_PREFIX)/2:len(BADGE_CODE_PART_HEX_PREFIX)/2+PUBKEY_HASH_LEN], &chaincfg.RegressionNetParams)
	if err != nil {
		panic("ParseBadgeVoutScript 5")
	}

	value := int64(binary.LittleEndian.Uint64(script[(len(BADGE_CODE_PART_HEX_PREFIX)+HEX_PUBKEY_HASH_LEN+len(BADGE_CODE_PART_HEX_SUFFIX)+len(BADGE_DATA_PART_HEX_PRIFIX))/2:]))
	if value < 0 {
		panic("ParseBadgeVoutScript 6")
	}
	return address, value
}

func ToBecPubkey(Key *btcec.PublicKey) *bec.PublicKey {
	return (*bec.PublicKey)(Key.ToECDSA())
}

func AddVin(msgTx *wire.MsgTx, prehashStr string, preindex int, script []byte) {
	prehash, err := chainhash.NewHashFromStr(prehashStr)
	if err != nil {
		panic(err)
	}
	preOutPoint := wire.NewOutPoint(prehash, uint32(preindex))
	vin := wire.NewTxIn(preOutPoint, script, nil)
	msgTx.AddTxIn(vin)
}

func AddVout(msgTx *wire.MsgTx, pkScript []byte, amount int64) {
	vout := wire.NewTxOut(amount, pkScript)
	msgTx.AddTxOut(vout)
}

func GetP2PKHUnlockScript(msgTx *wire.MsgTx, index int, key *btcec.PrivateKey, lockScript []byte, value int64) []byte {
	sig := GetSig(msgTx, index, lockScript, uint64(value), txscript.SigHashAll|SigHashForkID, key)

	builder := txscript.NewScriptBuilder()

	pubbyte := key.PubKey().SerializeCompressed()

	b, err := builder.AddData(sig).AddData(pubbyte).Script()
	if err != nil {
		panic(err)
	}
	return b
}

func LoadDesc(path string) *scryptlib.Contract {
	desc, err := scryptlib.LoadDesc(path)
	if err != nil {
		panic(err)
	}

	contract, err := scryptlib.NewContractFromDesc(desc)
	if err != nil {
		panic(err)
	}
	return &contract
}
