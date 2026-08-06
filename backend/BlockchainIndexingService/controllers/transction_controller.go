package controllers

import (
	"BlockchainIndexingService/services"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type TransactionController struct {
	service *services.TransactionService
}

func NewTransactionController(service *services.TransactionService) *TransactionController {
	return &TransactionController{service}
}

func (controller *TransactionController) RegisterRoutes(r chi.Router) {
	r.Route("/tx", func(r chi.Router) {
		r.Get("/{txID}", controller.GetTransaction)
		r.Post("/broadcast", controller.BroadcastTransaction)
		r.Post("/decode", controller.DecodeTransaction)
	})
}

func (controller *TransactionController) GetTransaction(w http.ResponseWriter, req *http.Request) {

}

func (controller *TransactionController) BroadcastTransaction(w http.ResponseWriter, req *http.Request) {

}

func (controller *TransactionController) DecodeTransaction(w http.ResponseWriter, req *http.Request) {
}
