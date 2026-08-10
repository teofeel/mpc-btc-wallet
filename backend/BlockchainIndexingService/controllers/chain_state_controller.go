package controllers

import (
	"BlockchainIndexingService/errors"
	"BlockchainIndexingService/services"
	"BlockchainIndexingService/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
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
	hash := chi.URLParam(r, "hash")

	block, err := controller.service.GetBlockByHash(r.Context(), hash)
	if err != nil {
		errors.GlobalErrorHandler(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, block)
}

func (controller *ChainStateController) GetBlockByHeight(w http.ResponseWriter, r *http.Request) {
	height := chi.URLParam(r, "height")

	heightInt, err := strconv.Atoi(height)
	if err != nil {
		errors.GlobalErrorHandler(w, errors.ErrInvalidArgument)
		return
	}

	block, err := controller.service.GetBlockByHeight(r.Context(), heightInt)
	if err != nil {
		errors.GlobalErrorHandler(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, block)
}

func (controller *ChainStateController) GetChainStatus(w http.ResponseWriter, r *http.Request) {
	res, err := controller.service.GetChainStatus(r.Context())
	if err != nil {
		errors.GlobalErrorHandler(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, res)
}

func (controller *ChainStateController) GetFeeEstimation(w http.ResponseWriter, r *http.Request) {
	targetBlocksStr := r.URL.Query().Get("blocks")
	targetBlocks := int64(6)

	if targetBlocksStr != "" {
		parsedTarget, err := strconv.ParseInt(targetBlocksStr, 10, 64)

		if err != nil || parsedTarget <= 0 {
			errors.GlobalErrorHandler(w, errors.ErrInvalidArgument)
			return
		}

		targetBlocks = parsedTarget
	}

	res, err := controller.service.GetFeeEstimation(r.Context(), targetBlocks)
	if err != nil {
		errors.GlobalErrorHandler(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, res)
}
