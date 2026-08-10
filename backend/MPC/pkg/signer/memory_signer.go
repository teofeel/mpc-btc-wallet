package signer

import (
	"context"
	"fmt"
	"math/big"
	"mpc/pkg/network"
	"mpc/pkg/storage"

	"github.com/bnb-chain/tss-lib/v2/common"
	"github.com/bnb-chain/tss-lib/v2/ecdsa/signing"
	"github.com/bnb-chain/tss-lib/v2/tss"
)

type Signature struct {
	Signature  []byte
	R          []byte
	S          []byte
	RecoveryID int32
}

type Signer interface {
	Sign(ctx context.Context, walletID string, msgHashBytes []byte) (*Signature, error)
}

type MemorySigner struct {
	router           network.Router
	shareStore       storage.PartyShareStore
	ownPartyID       *tss.PartyID
	sortedPartiesIDs []*tss.PartyID
	threshold        int
}

func NewMemorySigner(router network.Router, shares storage.PartyShareStore, ownPartyID *tss.PartyID, allPartyIDs []*tss.PartyID, threshold int) *MemorySigner {
	return &MemorySigner{
		router:           router,
		shareStore:       shares,
		ownPartyID:       ownPartyID,
		sortedPartiesIDs: allPartyIDs,
		threshold:        threshold,
	}
}

func (s *MemorySigner) Sign(ctx context.Context, walletID string, msgHashBytes []byte) (*Signature, error) {
	sharesByShareID, err := s.shareStore.GetAll(walletID)
	if err != nil {
		return nil, err
	}

	var self *tss.PartyID
	for _, p := range s.sortedPartiesIDs {
		if p.Id == s.ownPartyID.Id {
			self = p
			break
		}
	}

	if self == nil {
		return nil, fmt.Errorf("not found own party")
	}

	msgHash := new(big.Int).SetBytes(msgHashBytes)
	peerCtx := tss.NewPeerContext(s.sortedPartiesIDs)

	outCh := make(chan tss.Message, len(s.sortedPartiesIDs)*4)
	endCh := make(chan *common.SignatureData, len(s.sortedPartiesIDs))
	errCh := make(chan *tss.Error, len(s.sortedPartiesIDs))

	parties := make([]tss.Party, 0, len(s.sortedPartiesIDs))
	for _, partyID := range s.sortedPartiesIDs {
		share, ok := sharesByShareID[partyID.KeyInt().String()]
		if !ok {
			return nil, fmt.Errorf("missing key share for party %s (%s) of wallet %s", partyID.Id, partyID.Moniker, walletID)
		}

		params := tss.NewParameters(tss.S256(), peerCtx, partyID, len(s.sortedPartiesIDs), s.threshold)
		party := signing.NewLocalParty(msgHash, params, *share, outCh, endCh)

		parties = append(parties, party)

		s.router.Register(partyID.Id, party, func(e *tss.Error) {
			select {
			case errCh <- e:
			default:

			}
		})
	}

	defer func() {
		for _, partyID := range s.sortedPartiesIDs {
			s.router.Unregister(partyID.Id)
		}
	}()

	go func() {
		for {
			select {
			case msg := <-outCh:
				s.router.Route(msg, msg.GetFrom())
			case <-ctx.Done():
				return
			}
		}
	}()

	for _, party := range parties {
		go func(p tss.Party) {
			if err := p.Start(); err != nil {
				select {
				case errCh <- err:
				default:
				}
			}
		}(party)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("signing session for wallet %s: %w", walletID, ctx.Err())
		case tssErr := <-errCh:
			return nil, fmt.Errorf("signing party err: %w", tssErr)
		case sigData := <-endCh:
			var recoveryID int32
			if len(sigData.SignatureRecovery) > 0 {
				recoveryID = int32(sigData.SignatureRecovery[0])
			}

			return &Signature{
				Signature:  sigData.Signature,
				R:          sigData.R,
				S:          sigData.S,
				RecoveryID: recoveryID,
			}, nil
		}
	}
}
