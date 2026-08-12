package utils

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
)

func GetNetwork(network string) (*chaincfg.Params, error) {
	switch strings.ToLower(strings.TrimSpace(network)) {
	case "mainnet", "livenet", "":
		return &chaincfg.MainNetParams, nil
	case "testnet", "testnet3":
		return &chaincfg.TestNet3Params, nil
	case "regtest":
		return &chaincfg.RegressionNetParams, nil
	case "simnet":
		return &chaincfg.SimNetParams, nil
	case "signet":
		return &chaincfg.SigNetParams, nil
	default:
		return nil, fmt.Errorf("unsupported bitcoin network: %q", network)
	}
}
