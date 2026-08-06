package models

type UTXO struct {
	TxID          string `json:"txid"`
	Vout          uint32 `json:"vout"`
	Address       string `json:"address"`
	AmountSats    int64  `json:"amount_sats"`
	BlockHeight   int64  `json:"block_height"`
	Spent         bool   `json:"-"`
	SpentTxID     string `json:"-"`
	Confirmations int64  `json:"confirmations"`
}
