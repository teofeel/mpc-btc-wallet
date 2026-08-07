package services

import (
	"BlockchainIndexingService/errors"
	"BlockchainIndexingService/models"
	"BlockchainIndexingService/repository"
	"BlockchainIndexingService/utils"
	"context"
	"github.com/btcsuite/btcd/rpcclient"
	"net/http"
)

type AddressService struct {
	rpcClient *rpcclient.Client
	store     repository.Store
}

func NewAddressService(rpcClient *rpcclient.Client, store repository.Store) *AddressService {
	return &AddressService{rpcClient: rpcClient, store: store}
}

func (service *AddressService) GetAddressUTXOs(ctx context.Context, addr string) ([]models.UTXO, error) {
	tip, err := service.rpcClient.GetBlockCount()
	if err != nil {
		return nil, errors.ErrInternalServer
	}

	utxos, err := service.store.UTXOsForAddress(ctx, addr, tip)
	if err != nil {
		return nil, errors.ErrInternalServer
	}

	return utxos, nil
}

func (service *AddressService) GetWalletBalance(ctx context.Context, addr string) (int64, error) {

	balance, err := service.store.BalanceForAddress(ctx, addr)
	if err != nil {
		return 0, errors.ErrInvalidAddress
	}

	return balance, nil
}

func (service *AddressService) GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	utils.WriteErr(w, http.StatusNotImplemented, "not yet implemented")
}
