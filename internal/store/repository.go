package store

import (
	"context"

	"github.com/kusakari/itsumo/internal/domain"
)

type Repository interface {
	// Trades
	ListTrades(ctx context.Context) ([]*domain.Trade, error)
	GetTrade(ctx context.Context, id string) (*domain.Trade, error)
	CreateTrade(ctx context.Context, t *domain.Trade) error
	CloseTrade(ctx context.Context, id string, exitPrice float64, exitDate, exitRationale interface{}) error

	// Events
	ListEvents(ctx context.Context, tradeID string) ([]*domain.Event, error)
	GetEvent(ctx context.Context, id string) (*domain.Event, error)
	CreateEvent(ctx context.Context, e *domain.Event) error

	// Votes — atomic increment, returns updated counts
	VoteTrade(ctx context.Context, tradeID, target, dir string) (*domain.VoteCount, error)
	VoteEvent(ctx context.Context, eventID, dir string) (*domain.VoteCount, error)

	// Prices
	GetLatestPrice(ctx context.Context, code string) (*domain.DailyPrice, error)
	ListPricesByCode(ctx context.Context, code string, limit int) ([]*domain.DailyPrice, error)

	// Companies
	UpsertCompanies(ctx context.Context, companies []*domain.Company) error
	ListCompanyIndustries(ctx context.Context) ([]string, error)
	ListCompanies(ctx context.Context, industry, keyword string, limit int) ([]*domain.Company, error)

	// Earnings
	GetNextEarningsAnnouncement(ctx context.Context, code string) (*domain.EarningsAnnouncement, error)
}
