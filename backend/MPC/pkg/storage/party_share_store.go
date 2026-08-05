package storage

import "github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"

type PartyShareStore interface {
	Save(walletID string, saveData *keygen.LocalPartySaveData) error
	Get(walletID string) (*keygen.LocalPartySaveData, error)
	Delete(walletID string) error
	Exists(walletID string) bool
}
