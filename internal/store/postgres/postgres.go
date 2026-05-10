package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

// New は標準 PostgreSQL 環境変数（PGHOST, PGUSER, PGDATABASE, PGPASSWORD 等）から接続します。
func New(ctx context.Context) (*Repo, error) {
	pool, err := pgxpool.New(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	return &Repo{pool: pool}, nil
}

func (r *Repo) Close() {
	r.pool.Close()
}
