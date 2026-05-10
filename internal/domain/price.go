package domain

import "time"

// DailyPrice は1銘柄1日の株価データです。
// Firestore のドキュメントID は "{YYYYMMDD}_{code}" 形式です。
type DailyPrice struct {
	Code     string    `firestore:"code"`
	Name     string    `firestore:"name"`
	Industry string    `firestore:"industry"`
	Sector   string    `firestore:"sector"`
	Date     time.Time `firestore:"date"`
	Open     float64   `firestore:"open"`
	High     float64   `firestore:"high"`
	Low      float64   `firestore:"low"`
	Close    float64   `firestore:"close"`
	Volume   float64   `firestore:"volume"`
	SavedAt  time.Time `firestore:"saved_at"`
}
