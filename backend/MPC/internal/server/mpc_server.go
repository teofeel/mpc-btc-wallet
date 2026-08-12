package server

import (
	"math/big"
	"mpc/internal/handlers"
	"mpc/pkg/dkg"
	"mpc/pkg/network"
	signer2 "mpc/pkg/signer"
	"mpc/pkg/storage"

	"github.com/bnb-chain/tss-lib/v2/tss"
)

type MpcServer interface {
	handlers.KeyGenHandler
	handlers.TransactionHandler
}

type MemoryMpcServer struct {
	handlers.KeyGenHandler
	handlers.TransactionHandler
}

func NewMemoryMpcServer(jobManager *dkg.JobManager, router network.Router, shareStore storage.PartyShareStore, threshold int) *MemoryMpcServer {
	ownPartyID, allPartyIDs := newClusterParties()
	signer := signer2.NewMemorySigner(router, shareStore, ownPartyID, allPartyIDs, threshold)

	return &MemoryMpcServer{
		KeyGenHandler:      handlers.NewMemoryKeyGenHandler(jobManager, router, shareStore, ownPartyID, allPartyIDs),
		TransactionHandler: handlers.NewMemoryTransactionHandler(signer, router, shareStore, ownPartyID, allPartyIDs),
	}
}

func newClusterParties() (*tss.PartyID, []*tss.PartyID) {
	pID1 := tss.NewPartyID("1", "Node-1", big.NewInt(10))
	pID2 := tss.NewPartyID("2", "Node-2", big.NewInt(20))
	pID3 := tss.NewPartyID("3", "Node-3", big.NewInt(30))

	return pID1, tss.SortPartyIDs([]*tss.PartyID{pID1, pID2, pID3})
}
