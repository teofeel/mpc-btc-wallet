package utils

import (
	"github.com/btcsuite/btcd/chaincfg/v2"
	"log"
	"strings"
)

func GetChainParams(network string) *chaincfg.Params {
	switch strings.ToLower(strings.TrimSpace(network)) {
	case "mainnet", "main":
		return &chaincfg.MainNetParams
	case "testnet", "testnet3":
		return &chaincfg.TestNet3Params
	case "sigonet", "testnet4":
		return &chaincfg.SigNetParams
	case "regtest", "regression":
		return &chaincfg.RegressionNetParams
	default:
		log.Printf("Unknown or missing network '%s', defaulting to regtest", network)
		return &chaincfg.RegressionNetParams
	}
}
