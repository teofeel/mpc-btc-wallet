package controllers

import (
	"BlockchainIndexingService/errors"
	"BlockchainIndexingService/services"
	"BlockchainIndexingService/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type AddressController struct {
	service *services.AddressService
}

func NewAddressController(service *services.AddressService) *AddressController {
	return &AddressController{service}
}

func (controller *AddressController) RegisterRoutes(r chi.Router) {
	r.Route("/address", func(r chi.Router) {
		r.Get("/{address}/utxos", controller.GetUTXOs)
		r.Get("/{address}/balance", controller.GetBalance)
		r.Get("/{address}/tx", controller.GetTransactions)
	})
}

func (controller *AddressController) GetUTXOs(w http.ResponseWriter, r *http.Request) {
	addr := chi.URLParam(r, "address")

	utxos, err := controller.service.GetAddressUTXOs(r.Context(), addr)
	if err != nil {
		errors.GlobalErrorHandler(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utxos)
}

func (controller *AddressController) GetBalance(w http.ResponseWriter, r *http.Request) {
	addr := chi.URLParam(r, "address")

	balance, err := controller.service.GetWalletBalance(r.Context(), addr)
	if err != nil {
		errors.GlobalErrorHandler(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int64{"confirmed_balance": balance})
}

func (controller *AddressController) GetTransactions(w http.ResponseWriter, r *http.Request) {
	controller.service.GetWalletTransactions(w, r)
}
