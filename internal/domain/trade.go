package domain

import (
	"fmt"
	"time"
)

type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

type VoteCount struct {
	Up   int `firestore:"up"`
	Down int `firestore:"down"`
}

type EntryFocus string

const (
	FocusPriceUp EntryFocus = "price_up"
	FocusValue   EntryFocus = "value"
	FocusMarket  EntryFocus = "market"
)

// EntryTagOptions は買い根拠タグの選択肢
var EntryTagOptions = []string{
	"割安感",
	"事業の可能性",
	"将来的な価格上昇",
	"地合・テクニカル",
	"業績改善",
	"配当・還元期待",
	"テーマ性",
	"チャート形状",
}

type Trade struct {
	ID             string     `firestore:"-"`
	Ticker         string     `firestore:"ticker"`
	Name           string     `firestore:"name"`
	Status         Status     `firestore:"status"`
	EntryFocus     EntryFocus `firestore:"entry_focus"`
	EntryTags      []string   `firestore:"entry_tags"`
	EntryPrice     float64    `firestore:"entry_price"`
	EntryDate      time.Time  `firestore:"entry_date"`
	EntryRationale string     `firestore:"entry_rationale"`
	EntryVotes     VoteCount  `firestore:"entry_votes"`
	ExitPrice      float64    `firestore:"exit_price"`
	ExitDate       time.Time  `firestore:"exit_date"`
	ExitRationale  string     `firestore:"exit_rationale"`
	ExitVotes      VoteCount  `firestore:"exit_votes"`
	CreatedAt      time.Time  `firestore:"created_at"`
	UpdatedAt      time.Time  `firestore:"updated_at"`
}

func (t *Trade) PnLPercent() float64 {
	if t.Status == StatusOpen || t.EntryPrice == 0 {
		return 0
	}
	return (t.ExitPrice - t.EntryPrice) / t.EntryPrice * 100
}

func (t *Trade) PnLStr() string {
	p := t.PnLPercent()
	if p >= 0 {
		return fmt.Sprintf("+%.1f%%", p)
	}
	return fmt.Sprintf("%.1f%%", p)
}

type EventKind string

const (
	KindNote     EventKind = "note"
	KindEarnings EventKind = "earnings"
	KindNews     EventKind = "news"
)

type Event struct {
	ID        string    `firestore:"-"`
	TradeID   string    `firestore:"trade_id"`
	Kind      EventKind `firestore:"kind"`
	Body      string    `firestore:"body"`
	URL       string    `firestore:"url"`
	EventDate time.Time `firestore:"event_date"`
	Votes     VoteCount `firestore:"votes"`
	CreatedAt time.Time `firestore:"created_at"`
}
