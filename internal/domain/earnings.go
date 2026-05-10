package domain

import "time"

// FinancialStatement は1銘柄1決算期の財務データです。
// Firestore のドキュメントID は "{code}_{disclosed_YYYYMMDD}" 形式です。
type FinancialStatement struct {
	Code            string    `firestore:"code"`
	DisclosedDate   time.Time `firestore:"disclosed_date"`
	TypeOfDocument  string    `firestore:"type_of_document"` // Q1FY, Q2FY, Q3FY, FY 等
	PeriodStart     time.Time `firestore:"period_start"`
	PeriodEnd       time.Time `firestore:"period_end"`
	FiscalYearStart time.Time `firestore:"fiscal_year_start"`
	FiscalYearEnd   time.Time `firestore:"fiscal_year_end"`

	// 実績
	NetSales        *float64 `firestore:"net_sales"`
	OperatingProfit *float64 `firestore:"operating_profit"`
	OrdinaryProfit  *float64 `firestore:"ordinary_profit"`
	Profit          *float64 `firestore:"profit"`
	EPS             *float64 `firestore:"eps"`
	BPS             *float64 `firestore:"bps"`

	// 予想（会社ガイダンス）
	ForecastNetSales        *float64 `firestore:"forecast_net_sales"`
	ForecastOperatingProfit *float64 `firestore:"forecast_operating_profit"`
	ForecastProfit          *float64 `firestore:"forecast_profit"`
	ForecastEPS             *float64 `firestore:"forecast_eps"`

	SavedAt time.Time `firestore:"saved_at"`
}

// EarningsAnnouncement は決算発表予定です。
// Firestore のドキュメントID は "{code}_{date_YYYYMMDD}" 形式です。
type EarningsAnnouncement struct {
	Code          string    `firestore:"code"`
	CompanyName   string    `firestore:"company_name"`
	Date          time.Time `firestore:"date"`
	FiscalYear    string    `firestore:"fiscal_year"`
	FiscalQuarter string    `firestore:"fiscal_quarter"`
	SavedAt       time.Time `firestore:"saved_at"`
}
