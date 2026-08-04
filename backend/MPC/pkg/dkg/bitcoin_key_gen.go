package dkg

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"math/big"
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

	PreParams []*keygen.LocalPreParams
	Timeout   time.Duration
}

type KeyGenResult struct {
	WalletId       string
	PublicKeyHex   string
	ChainCodeHex   string
	XPub           string
	BitcoinAddress string
	SaveData       *keygen.LocalPartySaveData
}

func checkErrors(ctx context.Context, cfg KeyGenConfig) error {
	if cfg.Threshold >= len(cfg.AllPartyIDs) {
		return fmt.Errorf("Threshold %d must be less than party cound %d", cfg.Threshold, len(cfg.AllPartyIDs))
	}
	if len(cfg.PreParams) != len(cfg.AllPartyIDs) {
		return fmt.Errorf("Key gen for %s: pre-params must be generated per party", cfg.WalletID)
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

	outCh := make(chan tss.Message, len(cfg.AllPartyIDs)*4)
	endCh := make(chan *keygen.LocalPartySaveData, len(cfg.AllPartyIDs))
	errCh := make(chan *tss.Error, len(cfg.AllPartyIDs))

	parties := generateAndRegisterParties(cfg, peerCtx, outCh, endCh, errCh)

	defer func() {
		for _, partyID := range cfg.AllPartyIDs {
			cfg.Router.Unregister(partyID.Id)
		}
	}()

	go handleMessage(ctx, cfg, outCh)

	startParties(parties, errCh)

	select {
	case saveData := <-endCh:
		return buildResult(cfg.WalletID, saveData, cfg.Network)
	case tssErr := <-errCh:
		return nil, fmt.Errorf("%w", tssErr)
	case <-ctx.Done():
		return nil, fmt.Errorf("%w", ctx.Err())
	}
}

func generateAndRegisterParties(cfg KeyGenConfig, peerCtx *tss.PeerContext,
	outCh chan tss.Message, endCh chan *keygen.LocalPartySaveData, errCh chan *tss.Error) []tss.Party {

	parties := make([]tss.Party, 0, len(cfg.AllPartyIDs))
	for i, partyID := range cfg.AllPartyIDs {
		params := tss.NewParameters(tss.S256(), peerCtx, partyID, len(cfg.AllPartyIDs), cfg.Threshold)
		party := keygen.NewLocalParty(params, outCh, endCh, *cfg.PreParams[i])

		parties = append(parties, party)

		cfg.Router.Register(partyID.Id, party, func(err *tss.Error) {
			select {
			case errCh <- err:
			default:
			}
		})
	}

	return parties
}

func handleMessage(ctx context.Context, cfg KeyGenConfig, outCh chan tss.Message) {
	for {
		select {
		case msg := <-outCh:
			cfg.Router.Route(msg, msg.GetFrom())
		case <-ctx.Done():
			return
		}
	}
}

func startParties(parties []tss.Party, errCh chan *tss.Error) {
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

	xPub, chainCode, err := buildXPubFromTss(saveData, netParams)
	if err != nil {
		return nil, err
	}

	return &KeyGenResult{
		WalletId:       walletID,
		PublicKeyHex:   hex.EncodeToString(compressedPubKey),
		BitcoinAddress: address,
		ChainCodeHex:   chainCode,
		XPub:           xPub,
		SaveData:       saveData,
	}, nil
}

func buildXPubFromTss(saveData *keygen.LocalPartySaveData, netParams *chaincfg.Params) (string, string, error) {
	pkX := saveData.ECDSAPub.X()
	pkY := saveData.ECDSAPub.Y()

	pubKeyBytes := make([]byte, 33)
	if new(big.Int).And(pkY, big.NewInt(1)).Int64() == 0 {
		pubKeyBytes[0] = 0x02
	} else {
		pubKeyBytes[0] = 0x03
	}

	xBytes := pkX.Bytes()
	copy(pubKeyBytes[33-len(xBytes):], xBytes)

	h := sha256.New()
	h.Write(pubKeyBytes)
	chainCode := h.Sum(nil)

	xPub := hdkeychain.NewExtendedKey(
		netParams.HDPublicKeyID[:],
		pubKeyBytes,
		chainCode,
		[]byte{0x00, 0x00, 0x00, 0x00},
		0,
		0,
		false,
	)

	return xPub.String(), hex.EncodeToString(chainCode), nil
}

func getAddress(compressedPubKey []byte, netParams *chaincfg.Params) (string, error) {
	pubKeyHash := btcutil.Hash160(compressedPubKey)

	address, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, netParams)
	if err != nil {
		return "", err
	}

	return address.EncodeAddress(), nil
}
