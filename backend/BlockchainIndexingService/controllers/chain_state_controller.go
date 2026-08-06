package controllers

import (
	"BlockchainIndexingService/services"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type ChainStateController struct {
	service *services.ChainStateService
}

func NewChainStateController(service *services.ChainStateService) *ChainStateController {
	return &ChainStateController{service}
}

func (controller *ChainStateController) RegisterRoutes(r chi.Router) {
	r.Route("/block", func(r chi.Router) {
		r.Get("/hash/{hash}", controller.GetBlockByHash)
		r.Get("/height/{height}", controller.GetBlockByHeight)
	})

	r.Get("/chain-status", controller.GetChainStatus)
	r.Get("/fees/estimate", controller.GetFeeEstimation)
}

func (controller *ChainStateController) GetBlockByHash(w http.ResponseWriter, r *http.Request) {
	controller.service.GetBlockByHash(w, r)
}

func (controller *ChainStateController) GetBlockByHeight(w http.ResponseWriter, r *http.Request) {
	controller.service.GetBlockByHeight(w, r)
}

func (controller *ChainStateController) GetChainStatus(w http.ResponseWriter, r *http.Request) {
	controller.service.GetChainStatus(w, r)
}

func (controller *ChainStateController) GetFeeEstimation(w http.ResponseWriter, r *http.Request) {
	controller.service.GetFeeEstimation(w, r)
}
