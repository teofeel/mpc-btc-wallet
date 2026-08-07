package services

import (
	"BlockchainIndexingService/errors"
	"BlockchainIndexingService/repository"
	"BlockchainIndexingService/utils"
	"context"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chainhash/v2"
	"github.com/btcsuite/btcd/rpcclient"
)

type TransactionService struct {
	rpcClient *rpcclient.Client
	store     repository.Store
}

func NewTransactionService(rpcClient *rpcclient.Client, store repository.Store) *TransactionService {
	return &TransactionService{rpcClient: rpcClient, store: store}
}

func (service *TransactionService) GetTransaction(ctx context.Context, txID string) (*btcjson.GetTransactionResult, error) {
	hash, err := chainhash.NewHashFromStr(txID)
	if err != nil {
		return nil, errors.ErrInvalidTransaction
	}

	tx, err := service.rpcClient.GetTransaction(hash)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	return tx, nil
}

func (service *TransactionService) BroadcastTransaction(ctx context.Context, hexTx string) (string, error) {
	raw, err := hex.DecodeString(hexTx)
	if err != nil {
		return "", errors.ErrInvalidTransaction
	}

	msgTx, err := utils.DecodeTxBytes(raw)
	if err != nil {
		return "", errors.ErrInvalidTransaction
	}

	txHash, err := service.rpcClient.SendRawTransaction(msgTx, false)
	if err != nil {
		return "", errors.ErrInternalServer
	}

	return txHash.String(), nil
}

func (service *TransactionService) DecodeTransaction(ctx context.Context, hexTx string) (*btcjson.TxRawResult, error) {
	raw, err := hex.DecodeString(hexTx)
	if err != nil {
		return nil, errors.ErrInvalidTransaction
	}

	result, err := service.rpcClient.DecodeRawTransaction(raw)
	if err != nil {
		return nil, errors.ErrInvalidTransaction
	}

	return result, nil
}
