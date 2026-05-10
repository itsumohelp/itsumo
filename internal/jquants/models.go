package jquants

// DailyQuote は株価四本値レスポンス1件分です。
// v2 (/equities/bars/daily) と旧形式の両方を吸収します。
type DailyQuote struct {
	Date string `json:"Date"`
	Code string `json:"Code"`

	// v2 fields
	O  *float64 `json:"O"`
	H  *float64 `json:"H"`
	L  *float64 `json:"L"`
	C  *float64 `json:"C"`
	Vo *float64 `json:"Vo"`

	// legacy fields
	Open   *float64 `json:"Open"`
	High   *float64 `json:"High"`
	Low    *float64 `json:"Low"`
	Close  *float64 `json:"Close"`
	Volume *float64 `json:"Volume"`
}

func (q DailyQuote) OpenValue() float64 {
	if q.O != nil {
		return *q.O
	}
	if q.Open != nil {
		return *q.Open
	}
	return 0
}

func (q DailyQuote) HighValue() float64 {
	if q.H != nil {
		return *q.H
	}
	if q.High != nil {
		return *q.High
	}
	return 0
}

func (q DailyQuote) LowValue() float64 {
	if q.L != nil {
		return *q.L
	}
	if q.Low != nil {
		return *q.Low
	}
	return 0
}

func (q DailyQuote) CloseValue() float64 {
	if q.C != nil {
		return *q.C
	}
	if q.Close != nil {
		return *q.Close
	}
	return 0
}

func (q DailyQuote) VolumeValue() float64 {
	if q.Vo != nil {
		return *q.Vo
	}
	if q.Volume != nil {
		return *q.Volume
	}
	return 0
}

// EquityMaster は /equities/master のレスポンス1件分です。
type EquityMaster struct {
	Date        string `json:"Date"`
	Code        string `json:"Code"`
	CoName      string `json:"CoName"`
	CompanyName string `json:"CompanyName"`
	S17Nm       string `json:"S17Nm"`
	S33Nm       string `json:"S33Nm"`
}

func (m EquityMaster) Name() string {
	if m.CoName != "" {
		return m.CoName
	}
	return m.CompanyName
}

// FinancialStatement は /fins/statements のレスポンス1件分です。
type FinancialStatement struct {
	DisclosedDate              string   `json:"DisclosedDate"`
	DisclosedTime              string   `json:"DisclosedTime"`
	LocalCode                  string   `json:"LocalCode"`
	TypeOfDocument             string   `json:"TypeOfDocument"`
	TypeOfCurrentPeriod        string   `json:"TypeOfCurrentPeriod"`
	CurrentPeriodStartDate     string   `json:"CurrentPeriodStartDate"`
	CurrentPeriodEndDate       string   `json:"CurrentPeriodEndDate"`
	CurrentFiscalYearStartDate string   `json:"CurrentFiscalYearStartDate"`
	CurrentFiscalYearEndDate   string   `json:"CurrentFiscalYearEndDate"`
	NetSales                   *float64 `json:"NetSales"`
	OperatingProfit            *float64 `json:"OperatingProfit"`
	OrdinaryProfit             *float64 `json:"OrdinaryProfit"`
	Profit                     *float64 `json:"Profit"`
	EarningsPerShare           *float64 `json:"EarningsPerShare"`
	BookValuePerShare          *float64 `json:"BookValuePerShare"`
	ForecastNetSales           *float64 `json:"ForecastNetSales"`
	ForecastOperatingProfit    *float64 `json:"ForecastOperatingProfit"`
	ForecastProfit             *float64 `json:"ForecastProfit"`
	ForecastEarningsPerShare   *float64 `json:"ForecastEarningsPerShare"`
}

// EarningsAnnouncement は /fins/earnings_announcement_dates_times のレスポンス1件分です。
type EarningsAnnouncement struct {
	Code              string `json:"Code"`
	ScheduledDate     string `json:"ScheduledDate"`
	PublicationDate   string `json:"PublicationDate"`
	CompanyName       string `json:"CompanyName"`
	FiscalQuarterName string `json:"FiscalQuarterName"`
}
