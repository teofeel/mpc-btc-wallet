package storage

import "github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"

type PartyShareStore interface {
	Save(walletID string, saveData *keygen.LocalPartySaveData) error
	SaveAll(walletID string, saveDataByShare map[string]*keygen.LocalPartySaveData) error
	Get(walletID string) (*keygen.LocalPartySaveData, error)
	GetAll(walletID string) (map[string]*keygen.LocalPartySaveData, error)
	Delete(walletID string) error
	Exists(walletID string) bool
}
