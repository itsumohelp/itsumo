package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kusakari/itsumo/internal/domain"
)

const priceBatchSize = 500

const upsertPriceSQL = `
INSERT INTO daily_prices (code, date, name, industry, sector, open, high, low, close, volume, saved_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (code, date) DO UPDATE SET
    name     = EXCLUDED.name,
    industry = EXCLUDED.industry,
    sector   = EXCLUDED.sector,
    open     = EXCLUDED.open,
    high     = EXCLUDED.high,
    low      = EXCLUDED.low,
    close    = EXCLUDED.close,
    volume   = EXCLUDED.volume,
    saved_at = EXCLUDED.saved_at`

// SaveDailyPrices は全銘柄の株価を PostgreSQL にバッチ upsert します。冪等です。
func (r *Repo) SaveDailyPrices(ctx context.Context, prices []*domain.DailyPrice) error {
	now := time.Now()
	for i := 0; i < len(prices); i += priceBatchSize {
		end := i + priceBatchSize
		if end > len(prices) {
			end = len(prices)
		}
		chunk := prices[i:end]

		batch := &pgx.Batch{}
		for _, p := range chunk {
			batch.Queue(upsertPriceSQL,
				p.Code, p.Date, p.Name, p.Industry, p.Sector,
				p.Open, p.High, p.Low, p.Close, p.Volume, now,
			)
		}

		br := r.pool.SendBatch(ctx, batch)
		for range chunk {
			if _, err := br.Exec(); err != nil {
				br.Close()
				return fmt.Errorf("prices batch (offset %d): %w", i, err)
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("prices batch close (offset %d): %w", i, err)
		}
	}
	return nil
}

// ListPricesByCode は指定銘柄の直近 N 件を日付降順で返します。
func (r *Repo) ListPricesByCode(ctx context.Context, code string, limit int) ([]*domain.DailyPrice, error) {
	const q = `
		SELECT code, date, name, industry, sector, open, high, low, close, volume, saved_at
		FROM daily_prices
		WHERE code = $1
		ORDER BY date DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, q, code, limit)
	if err != nil {
		return nil, fmt.Errorf("ListPricesByCode: %w", err)
	}
	defer rows.Close()

	var prices []*domain.DailyPrice
	for rows.Next() {
		p := &domain.DailyPrice{}
		if err := rows.Scan(
			&p.Code, &p.Date, &p.Name, &p.Industry, &p.Sector,
			&p.Open, &p.High, &p.Low, &p.Close, &p.Volume, &p.SavedAt,
		); err != nil {
			return nil, fmt.Errorf("ListPricesByCode scan: %w", err)
		}
		prices = append(prices, p)
	}
	return prices, rows.Err()
}

// GetLatestPrice は指定銘柄の最新株価を返します。データがない場合は nil, nil を返します。
func (r *Repo) GetLatestPrice(ctx context.Context, code string) (*domain.DailyPrice, error) {
	const q = `
		SELECT code, date, name, industry, sector, open, high, low, close, volume, saved_at
		FROM daily_prices
		WHERE code = $1
		ORDER BY date DESC
		LIMIT 1`

	p := &domain.DailyPrice{}
	err := r.pool.QueryRow(ctx, q, code).Scan(
		&p.Code, &p.Date, &p.Name, &p.Industry, &p.Sector,
		&p.Open, &p.High, &p.Low, &p.Close, &p.Volume, &p.SavedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetLatestPrice: %w", err)
	}
	return p, nil
}
