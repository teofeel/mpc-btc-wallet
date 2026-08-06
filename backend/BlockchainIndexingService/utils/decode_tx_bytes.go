package utils

import (
	"bytes"
	"github.com/btcsuite/btcd/wire/v2"
)

func DecodeTxBytes(raw []byte) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(wire.TxVersion)
	if err := tx.Deserialize(bytes.NewReader(raw)); err != nil {
		return nil, err
	}

	return tx, nil
}
