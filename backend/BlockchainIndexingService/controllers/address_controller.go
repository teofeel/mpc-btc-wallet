package controllers

import (
	"BlockchainIndexingService/services"
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
	controller.service.GetAddressUTXOs(w, r)
}

func (controller *AddressController) GetBalance(w http.ResponseWriter, r *http.Request) {
	controller.service.GetWalletBalance(w, r)
}

func (controller *AddressController) GetTransactions(w http.ResponseWriter, r *http.Request) {
	controller.service.GetWalletTransactions(w, r)
}
