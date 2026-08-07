package controllers

import (
	"BlockchainIndexingService/dtos"
	"BlockchainIndexingService/errors"
	"BlockchainIndexingService/services"
	"BlockchainIndexingService/utils"
	"encoding/json"
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

func (controller *TransactionController) GetTransaction(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txID")

	tx, err := controller.service.GetTransaction(r.Context(), txID)
	if err != nil {
		errors.GlobalErrorHandler(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, tx)
}

func (controller *TransactionController) BroadcastTransaction(w http.ResponseWriter, r *http.Request) {
	var body dtos.BroadcastTxRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	txID, err := controller.service.BroadcastTransaction(r.Context(), body.Hex)
	if err != nil {
		errors.GlobalErrorHandler(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"txID": txID})
}

func (controller *TransactionController) DecodeTransaction(w http.ResponseWriter, r *http.Request) {
	var body dtos.DecodeTxRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := controller.service.DecodeTransaction(r.Context(), body.Hex)
	if err != nil {
		errors.GlobalErrorHandler(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, result)
}
