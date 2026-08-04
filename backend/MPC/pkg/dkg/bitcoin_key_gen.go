package dkg

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"mpc/pkg/network"
	"time"
)

type KeyGenConfig struct {
	WalletID    string
	PartyID     *tss.PartyID
	AllPartyIDs tss.SortedPartyIDs
	Threshold   int
	Router      network.Router
	Network     *chaincfg.Params

	PreParams *keygen.LocalPreParams
	Timeout   time.Duration
}

type KeyGenResult struct {
	WalletId       string
	PublicKeyHex   string
	BitcoinAddress string
	SaveData       *keygen.LocalPartySaveData
}

func checkErrors(ctx context.Context, cfg KeyGenConfig) error {
	if cfg.Threshold >= len(cfg.AllPartyIDs) {
		return fmt.Errorf("Threshold %d must be less than party cound &d", cfg.Threshold, len(cfg.AllPartyIDs))
	}
	if cfg.PreParams == nil {
		return fmt.Errorf("Key gen for %s: no pre generated params", cfg.WalletID)
	}

	return nil
}

func RunKeyGen(ctx context.Context, cfg KeyGenConfig) (*KeyGenResult, error) {
	err := checkErrors(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 2 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	peerCtx := tss.NewPeerContext(cfg.AllPartyIDs)
	params := tss.NewParameters(tss.S256(), peerCtx, cfg.PartyID, len(cfg.AllPartyIDs), cfg.Threshold)

	outCh := make(chan tss.Message, len(cfg.AllPartyIDs))
	endCh := make(chan *keygen.LocalPartySaveData, 1)
	errCh := make(chan *tss.Error, 1)

	party := keygen.NewLocalParty(params, outCh, endCh, *cfg.PreParams)

	cfg.Router.Register(cfg.PartyID.Id, party, func(err *tss.Error) {
		select {
		case errCh <- err:
		default:
		}
	})

	defer cfg.Router.Unregister(cfg.PartyID.Id)

	go func() {
		if err := party.Start(); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-outCh:
				cfg.Router.Route(msg, cfg.PartyID)
			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case saveData := <-endCh:
		return buildResult(cfg.WalletID, saveData, cfg.Network)
	case tssErr := <-errCh:
		return nil, fmt.Errorf("%w", tssErr)
	case <-ctx.Done():
		return nil, fmt.Errorf("%w", ctx.Err())
	}
}

func buildResult(walletID string, saveData *keygen.LocalPartySaveData, netParams *chaincfg.Params) (*KeyGenResult, error) {
	ecPoint := saveData.ECDSAPub
	if ecPoint == nil {
		return nil, fmt.Errorf("Couldnt generate key for wallet %s", walletID)
	}

	pubKey := &ecdsa.PublicKey{
		Curve: tss.S256(),
		X:     ecPoint.X(),
		Y:     ecPoint.Y(),
	}

	compressedPubKey := elliptic.MarshalCompressed(pubKey.Curve, pubKey.X, pubKey.Y)

	address, err := getAddress(compressedPubKey, netParams)
	if err != nil {
		return nil, fmt.Errorf("wallet %s failed to derive address %w", walletID, err)
	}

	return &KeyGenResult{
		WalletId:       walletID,
		PublicKeyHex:   hex.EncodeToString(compressedPubKey),
		BitcoinAddress: address,
		SaveData:       saveData,
	}, nil
}

func getAddress(compressedPubKey []byte, netParams *chaincfg.Params) (string, error) {
	pubKeyHash := btcutil.Hash160(compressedPubKey)

	address, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, netParams)
	if err != nil {
		return "", err
	}

	return address.EncodeAddress(), nil
}
