package firestore

import (
	"context"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/kusakari/itsumo/internal/domain"
)

type Repo struct {
	client *fs.Client
}

func New(ctx context.Context, projectID string) (*Repo, error) {
	client, err := fs.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &Repo{client: client}, nil
}

func (r *Repo) Close() error {
	return r.client.Close()
}

// ── Trades ────────────────────────────────────────────────────────────────────

func (r *Repo) ListTrades(ctx context.Context) ([]*domain.Trade, error) {
	docs, err := r.client.Collection("trades").
		OrderBy("created_at", fs.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	trades := make([]*domain.Trade, 0, len(docs))
	for _, doc := range docs {
		t := &domain.Trade{}
		if err := doc.DataTo(t); err != nil {
			return nil, err
		}
		t.ID = doc.Ref.ID
		trades = append(trades, t)
	}
	return trades, nil
}

func (r *Repo) GetTrade(ctx context.Context, id string) (*domain.Trade, error) {
	doc, err := r.client.Collection("trades").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	t := &domain.Trade{}
	if err := doc.DataTo(t); err != nil {
		return nil, err
	}
	t.ID = doc.Ref.ID
	return t, nil
}

func (r *Repo) CreateTrade(ctx context.Context, t *domain.Trade) error {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	ref := r.client.Collection("trades").NewDoc()
	t.ID = ref.ID
	_, err := ref.Set(ctx, t)
	return err
}

func (r *Repo) CloseTrade(ctx context.Context, id string, exitPrice float64, exitDate, exitRationale interface{}) error {
	now := time.Now()
	_, err := r.client.Collection("trades").Doc(id).Update(ctx, []fs.Update{
		{Path: "status", Value: string(domain.StatusClosed)},
		{Path: "exit_price", Value: exitPrice},
		{Path: "exit_date", Value: exitDate},
		{Path: "exit_rationale", Value: exitRationale},
		{Path: "updated_at", Value: now},
	})
	return err
}

// ── Events ────────────────────────────────────────────────────────────────────

func (r *Repo) ListEvents(ctx context.Context, tradeID string) ([]*domain.Event, error) {
	docs, err := r.client.Collection("events").
		Where("trade_id", "==", tradeID).
		OrderBy("event_date", fs.Asc).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	events := make([]*domain.Event, 0, len(docs))
	for _, doc := range docs {
		e := &domain.Event{}
		if err := doc.DataTo(e); err != nil {
			return nil, err
		}
		e.ID = doc.Ref.ID
		events = append(events, e)
	}
	return events, nil
}

func (r *Repo) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	doc, err := r.client.Collection("events").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	e := &domain.Event{}
	if err := doc.DataTo(e); err != nil {
		return nil, err
	}
	e.ID = doc.Ref.ID
	return e, nil
}

func (r *Repo) CreateEvent(ctx context.Context, e *domain.Event) error {
	e.CreatedAt = time.Now()
	ref := r.client.Collection("events").NewDoc()
	e.ID = ref.ID
	_, err := ref.Set(ctx, e)
	return err
}

// ── Votes ─────────────────────────────────────────────────────────────────────

func (r *Repo) VoteTrade(ctx context.Context, tradeID, target, dir string) (*domain.VoteCount, error) {
	field := target + "_votes." + dir
	ref := r.client.Collection("trades").Doc(tradeID)
	_, err := ref.Update(ctx, []fs.Update{
		{Path: field, Value: fs.Increment(1)},
	})
	if err != nil {
		return nil, err
	}
	t, err := r.GetTrade(ctx, tradeID)
	if err != nil {
		return nil, err
	}
	if target == "entry" {
		return &t.EntryVotes, nil
	}
	return &t.ExitVotes, nil
}

func (r *Repo) VoteEvent(ctx context.Context, eventID, dir string) (*domain.VoteCount, error) {
	field := "votes." + dir
	ref := r.client.Collection("events").Doc(eventID)
	_, err := ref.Update(ctx, []fs.Update{
		{Path: field, Value: fs.Increment(1)},
	})
	if err != nil {
		return nil, err
	}
	e, err := r.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return &e.Votes, nil
}
