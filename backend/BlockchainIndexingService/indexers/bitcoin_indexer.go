package indexers

import (
	"BlockchainIndexingService/models"
	"BlockchainIndexingService/repository"
	"BlockchainIndexingService/utils"
	"context"
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg/v2"
	"github.com/btcsuite/btcd/chainhash/v2"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript/v2"
)

type BitcoinIndexer struct {
	rpcClient *rpcclient.Client
	store     repository.Store
	net       *chaincfg.Params
}

func NewIndexer(rpc *rpcclient.Client, store repository.Store) *BitcoinIndexer {
	networkEnv := utils.GetEnv("BTC_NETWORK", "regtest")
	params := utils.GetChainParams(networkEnv)

	return &BitcoinIndexer{
		rpcClient: rpc,
		store:     store,
		net:       params,
	}
}

func (idx *BitcoinIndexer) Run(ctx context.Context) {
	for {
		if err := idx.catchUp(ctx); err != nil {
			log.Println(err)
		}

		time.Sleep(10 * time.Second)
	}
}

func (idx *BitcoinIndexer) catchUp(ctx context.Context) error {
	log.Println("Catching up")
	tip, err := idx.rpcClient.GetBlockCount()
	if err != nil {
		return err
	}

	last, err := idx.store.LastIndexedHeight(ctx)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] DB Last Height: %d | Bitcoind Tip Height: %d\n", last, tip)

	for h := last + 1; h <= tip; h++ {
		if err := idx.indexBlockAtHeight(ctx, h); err != nil {
			return err
		}

		if err := idx.store.SetLastIndexedHeight(ctx, h); err != nil {
			return err
		}

		if h%1000 == 0 {
			log.Printf("indexed height %d / %d\n", h, tip)
		}
	}

	return nil
}

func (idx *BitcoinIndexer) indexBlockAtHeight(ctx context.Context, height int64) error {
	hash, err := idx.rpcClient.GetBlockHash(height)
	if err != nil {
		return err
	}

	block, err := idx.rpcClient.GetBlock(hash)
	if err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		txID := tx.TxHash().String()

		for _, in := range tx.TxIn {
			if in.PreviousOutPoint.Hash == (chainhash.Hash{}) {
				continue
			}

			_ = idx.store.MarkSpent(ctx, in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index, txID)
		}

		for voutIdx, out := range tx.TxOut {
			_, addrs, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, idx.net)
			if err != nil || len(addrs) == 0 {
				continue
			}

			for _, addr := range addrs {
				utxo := models.UTXO{
					TxID:        txID,
					Vout:        uint32(voutIdx),
					Address:     addr.EncodeAddress(),
					AmountSats:  out.Value,
					BlockHeight: height,
				}

				if err := idx.store.InsertUTXO(ctx, utxo); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
