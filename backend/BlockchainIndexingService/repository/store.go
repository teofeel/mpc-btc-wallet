package repository

import (
	"BlockchainIndexingService/models"
	"context"
)

type Store interface {
	InitSchema(ctx context.Context) error
	LastIndexedHeight(ctx context.Context) (int64, error)
	SetLastIndexedHeight(ctx context.Context, height int64) error
	InsertUTXO(ctx context.Context, u models.UTXO) error
	MarkSpent(ctx context.Context, txid string, vout uint32, spendingTxID string) error
	UTXOsForAddress(ctx context.Context, address string, tipHeight int64) ([]models.UTXO, error)
	BalanceForAddress(ctx context.Context, address string) (int64, error)
}
