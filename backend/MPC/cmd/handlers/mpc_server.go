package handlers

import (
	"fmt"
	"math/big"
	"mpc/pkg/dkg"
	"mpc/pkg/network"
	signer2 "mpc/pkg/signer"
	"mpc/pkg/storage"
	"strings"

	"github.com/bnb-chain/tss-lib/v2/tss"
	"github.com/btcsuite/btcd/chaincfg"
)

type MpcServer interface {
	KeyGenHandler
	TransactionHandler
}

type MemoryMpcServer struct {
	KeyGenHandler
	TransactionHandler
}

func NewMemoryMpcServer(jobManager *dkg.JobManager, router network.Router, shareStore storage.PartyShareStore, threshold int) *MemoryMpcServer {
	ownPartyID, allPartyIDs := newClusterParties()
	signer := signer2.NewMemorySigner(router, shareStore, ownPartyID, allPartyIDs, threshold)

	return &MemoryMpcServer{
		KeyGenHandler:      NewMemoryKeyGenHandler(jobManager, router, shareStore, ownPartyID, allPartyIDs),
		TransactionHandler: NewMemoryTransactionHandler(signer, router, shareStore, ownPartyID, allPartyIDs),
	}
}

func newClusterParties() (*tss.PartyID, []*tss.PartyID) {
	pID1 := tss.NewPartyID("1", "Node-1", big.NewInt(10))
	pID2 := tss.NewPartyID("2", "Node-2", big.NewInt(20))
	pID3 := tss.NewPartyID("3", "Node-3", big.NewInt(30))

	return pID1, tss.SortPartyIDs([]*tss.PartyID{pID1, pID2, pID3})
}

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
