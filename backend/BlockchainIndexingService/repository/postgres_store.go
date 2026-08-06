package repository

import (
	"BlockchainIndexingService/models"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool}
}
func (s *PostgresStore) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS utxos (
		txid TEXT NOT NULL,
		vout INT NOT NULL,
		address TEXT NOT NULL,
		amount_sats BIGINT NOT NULL,
		block_height BIGINT NOT NULL,
		spent BOOLEAN NOT NULL DEFAULT FALSE,
		spent_txid TEXT,
		PRIMARY KEY (txid, vout)
	);
	CREATE INDEX IF NOT EXISTS idx_utxos_address ON utxos(address);
	CREATE INDEX IF NOT EXISTS idx_utxos_spent ON utxos(spent);

	CREATE TABLE IF NOT EXISTS sync_state (
		id INT PRIMARY KEY CHECK (id = 1),
		last_indexed_height BIGINT NOT NULL DEFAULT -1
	);
	INSERT INTO sync_state (id, last_indexed_height) 
	VALUES (1, -1) 
	ON CONFLICT (id) DO NOTHING;
	`
	_, err := s.pool.Exec(ctx, schema)
	return err
}

func (s *PostgresStore) LastIndexedHeight(ctx context.Context) (int64, error) {
	var h int64
	err := s.pool.QueryRow(ctx, `SELECT last_indexed_height FROM sync_state WHERE id = 1`).Scan(&h)

	return h, err
}

func (s *PostgresStore) SetLastIndexedHeight(ctx context.Context, height int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE sync_state SET last_indexed_height = $1 WHERE id = 1`, height)
	return err
}

func (s *PostgresStore) InsertUTXO(ctx context.Context, u models.UTXO) error {
	query := `
		INSERT INTO utxos (txid, vout, address, amount_sats, block_height, spent)
		VALUES ($1, $2, $3, $4, $5, FALSE)
		ON CONFLICT (txid, vout) DO UPDATE SET
			address = EXCLUDED.address,
			amount_sats = EXCLUDED.amount_sats,
			block_height = EXCLUDED.block_height,
			spent = FALSE,
			spent_txid = NULL
	`
	_, err := s.pool.Exec(ctx, query, u.TxID, u.Vout, u.Address, u.AmountSats, u.BlockHeight)
	return err
}

func (s *PostgresStore) MarkSpent(ctx context.Context, txid string, vout uint32, spendingTxID string) error {
	query := `UPDATE utxos SET spent = TRUE, spent_txid = $1 WHERE txid = $2 AND vout = $3`
	_, err := s.pool.Exec(ctx, query, spendingTxID, txid, vout)
	return err
}

func (s *PostgresStore) UTXOsForAddress(ctx context.Context, address string, tipHeight int64) ([]models.UTXO, error) {
	query := `
		SELECT txid, vout, address, amount_sats, block_height
		FROM utxos 
		WHERE address = $1 AND spent = FALSE
	`
	rows, err := s.pool.Query(ctx, query, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.UTXO
	for rows.Next() {
		var u models.UTXO
		if err := rows.Scan(&u.TxID, &u.Vout, &u.Address, &u.AmountSats, &u.BlockHeight); err != nil {
			return nil, err
		}
		u.Confirmations = tipHeight - u.BlockHeight + 1
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *PostgresStore) BalanceForAddress(ctx context.Context, address string) (int64, error) {
	var total *int64
	query := `SELECT SUM(amount_sats) FROM utxos WHERE address = $1 AND spent = FALSE`
	err := s.pool.QueryRow(ctx, query, address).Scan(&total)
	if err != nil {
		return 0, err
	}
	if total == nil {
		return 0, nil
	}
	return *total, nil
}
