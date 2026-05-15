package postgres

import (
	"context"
	"fmt"

	"github.com/kusakari/itsumo/internal/domain"
)

func (r *Repo) SearchStocks(ctx context.Context, q string) ([]domain.StockSuggestion, error) {
	const sql = `
		SELECT DISTINCT ON (code) code, name
		FROM daily_prices
		WHERE code LIKE $1 OR name LIKE $2
		ORDER BY code, date DESC
		LIMIT 10`

	rows, err := r.pool.Query(ctx, sql, q+"%", "%"+q+"%")
	if err != nil {
		return nil, fmt.Errorf("SearchStocks: %w", err)
	}
	defer rows.Close()

	var results []domain.StockSuggestion
	for rows.Next() {
		var s domain.StockSuggestion
		if err := rows.Scan(&s.Code, &s.Name); err != nil {
			return nil, fmt.Errorf("SearchStocks scan: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

func (r *Repo) ListHoldings(ctx context.Context, userID string) ([]*domain.Holding, error) {
	const q = `
		SELECT h.user_id, h.stock_code, h.shares, h.updated_at,
		       COALESCE(p.name, '') AS name
		FROM holdings h
		LEFT JOIN LATERAL (
		    SELECT name FROM daily_prices WHERE code = h.stock_code ORDER BY date DESC LIMIT 1
		) p ON TRUE
		WHERE h.user_id = $1
		ORDER BY h.stock_code`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("ListHoldings: %w", err)
	}
	defer rows.Close()

	var holdings []*domain.Holding
	for rows.Next() {
		h := &domain.Holding{}
		if err := rows.Scan(&h.UserID, &h.StockCode, &h.Shares, &h.UpdatedAt, &h.StockName); err != nil {
			return nil, fmt.Errorf("ListHoldings scan: %w", err)
		}
		holdings = append(holdings, h)
	}
	return holdings, rows.Err()
}

func (r *Repo) UpsertHolding(ctx context.Context, userID, stockCode string, shares int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO holdings (user_id, stock_code, shares, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, stock_code) DO UPDATE SET shares = $3, updated_at = now()`,
		userID, stockCode, shares,
	)
	return err
}

func (r *Repo) DeleteHolding(ctx context.Context, userID, stockCode string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM holdings WHERE user_id = $1 AND stock_code = $2`,
		userID, stockCode,
	)
	return err
}
