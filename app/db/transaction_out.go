package db

import (
	"bytes"
	"git.jasonc.me/main/bitcoin/bitcoin/memo"
	"github.com/cpacia/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"strings"
	"time"
)

type TransactionOut struct {
	Id            uint   `gorm:"primary_key"`
	Index         uint32
	TransactionId uint   `gorm:"unique_index:transaction_out_script;"`
	Transaction   *Transaction
	Value         int64
	PkScript      []byte `gorm:"unique_index:transaction_out_script;"`
	LockString    string
	RequiredSigs  uint
	ScriptClass   uint
	Addresses     []*Address
	TxnInId       uint
	TxnIn         *TransactionIn
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (t TransactionOut) Save() error {
	result := save(&t)
	if result.Error != nil {
		return jerr.Get("error saving transaction output", result.Error)
	}
	return nil
}

func (t TransactionOut) ValueInBCH() float64 {
	return float64(t.Value) * 1.e-8
}

func (t TransactionOut) HasIn() bool {
	return t.TxnInId > 0
}

func (t TransactionOut) IsSpendable() bool {
	if t.TxnInId > 0 {
		txIn, _ := GetTransactionInputById(t.TxnInId)
		if txIn.Transaction.BlockId > 0 {
			return false
		}
	}
	transactionAddress := t.Transaction.Key.GetAddress().GetEncoded()
	var addressFound bool
	for _, address := range t.Addresses {
		if address.String == transactionAddress {
			addressFound = true
		}
	}
	return addressFound
}

func (t TransactionOut) GetScriptClass() string {
	return txscript.ScriptClass(t.ScriptClass).String()
}

func (t TransactionOut) GetMessage() string {
	if txscript.ScriptClass(t.ScriptClass) == txscript.NullDataTy {
		data, err := txscript.PushedData(t.PkScript)
		if err != nil || len(data) == 0 {
			return ""
		}
		return string(data[0])
	}
	if len(t.PkScript) < 5 || ! bytes.Equal(t.PkScript[0:3], []byte{
		txscript.OP_RETURN,
		txscript.OP_DATA_2,
		memo.CodePrefix,
	}) {
		return ""
	}
	data, err := txscript.PushedData(t.PkScript[3:])
	if err != nil || len(data) == 0 {
		return ""
	}
	var stringArray []string
	for _, bt := range data {
		stringArray = append(stringArray, string(bt))
	}
	return strings.Join(stringArray, " ")
}

func GetTransactionOutputById(id uint) (*TransactionOut, error) {
	query, err := getDb()
	if err != nil {
		return nil, jerr.Get("error getting db", err)
	}
	var txOut TransactionOut
	result := query.
		Preload("Transaction").
		Preload("Transaction.Key").
		Find(&txOut, TransactionOut{
		Id: id,
	})
	if result.Error != nil {
		return nil, jerr.Get("error finding transaction out", result.Error)
	}
	return &txOut, nil
}