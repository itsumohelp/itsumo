package domain

import "time"

type Holding struct {
	UserID    string
	StockCode string
	StockName string
	Shares    int
	UpdatedAt time.Time
}

type StockSuggestion struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
