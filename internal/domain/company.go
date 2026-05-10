package domain

import "time"

// Company は銘柄マスタの要約情報です。
type Company struct {
	Code      string    `firestore:"code"`
	Name      string    `firestore:"name"`
	Industry  string    `firestore:"industry"`
	Sector    string    `firestore:"sector"`
	UpdatedAt time.Time `firestore:"updated_at"`
}
