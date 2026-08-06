package services

import (
	"BlockchainIndexingService/repository"
	"github.com/btcsuite/btcd/rpcclient"
	"net/http"
)

type TransactionService struct {
	rpcClient *rpcclient.Client
	store     repository.Store
}

func NewTransactionService(rpcClient *rpcclient.Client, store repository.Store) *TransactionService {
	return &TransactionService{rpcClient: rpcClient, store: store}
}

func (service *TransactionService) GetTransaction(w http.ResponseWriter, req *http.Request) {

}

func (service *TransactionService) BroadcastTransaction(w http.ResponseWriter, req *http.Request) {

}

func (service *TransactionService) DecodeTransaction(w http.ResponseWriter, req *http.Request) {
	
}
