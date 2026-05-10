// fetch-earnings は J-Quants API から当日開示された決算情報と
// 直近の決算発表予定を取得し、Firestore に保存する Cloud Run Job です。
// Cloud Scheduler から毎営業日 22:00 JST に起動することを想定しています。
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/kusakari/itsumo/internal/domain"
	"github.com/kusakari/itsumo/internal/jquants"
	"github.com/kusakari/itsumo/internal/store/firestore"
)

func main() {
	ctx := context.Background()

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID is required")
	}

	date := targetDate()
	log.Printf("fetch-earnings: date=%s", date)

	jq := newJQuantsClient()

	repo, err := firestore.New(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore.New: %v", err)
	}
	defer repo.Close()

	if err := fetchStatements(ctx, jq, repo, date); err != nil {
		log.Fatalf("fetchStatements: %v", err)
	}
	if err := fetchAnnouncements(ctx, jq, repo); err != nil {
		log.Fatalf("fetchAnnouncements: %v", err)
	}

	log.Println("done.")
}

// fetchStatements は指定日に開示された決算情報を取得して保存します。
func fetchStatements(ctx context.Context, jq *jquants.Client, repo *firestore.Repo, date string) error {
	log.Printf("fetching financial statements disclosed on %s...", date)
	raw, err := jq.FinancialStatementsByDate(ctx, date)
	if err != nil {
		return err
	}
	log.Printf("fetched %d statements", len(raw))
	if len(raw) == 0 {
		return nil
	}

	stmts := make([]*domain.FinancialStatement, 0, len(raw))
	for _, s := range raw {
		stmt, err := convertStatement(s)
		if err != nil {
			log.Printf("skip statement code=%s: %v", s.LocalCode, err)
			continue
		}
		stmts = append(stmts, stmt)
	}

	log.Printf("saving %d statements to Firestore...", len(stmts))
	return repo.SaveFinancialStatements(ctx, stmts)
}

// fetchAnnouncements は直近の決算発表予定を取得して保存します。
func fetchAnnouncements(ctx context.Context, jq *jquants.Client, repo *firestore.Repo) error {
	log.Println("fetching earnings announcements...")
	raw, err := jq.EarningsAnnouncements(ctx)
	if err != nil {
		return err
	}
	log.Printf("fetched %d announcements", len(raw))
	if len(raw) == 0 {
		return nil
	}

	announcements := make([]*domain.EarningsAnnouncement, 0, len(raw))
	for _, a := range raw {
		d, err := time.ParseInLocation("2006-01-02", a.ScheduledDate, time.Local)
		if err != nil {
			log.Printf("skip announcement code=%s: invalid date %q", a.Code, a.ScheduledDate)
			continue
		}
		announcements = append(announcements, &domain.EarningsAnnouncement{
			Code:          a.Code,
			CompanyName:   a.CompanyName,
			Date:          d,
			FiscalQuarter: a.FiscalQuarterName,
		})
	}

	log.Printf("saving %d announcements to Firestore...", len(announcements))
	return repo.SaveEarningsAnnouncements(ctx, announcements)
}

func convertStatement(s jquants.FinancialStatement) (*domain.FinancialStatement, error) {
	disclosed, err := time.ParseInLocation("2006-01-02", s.DisclosedDate, time.Local)
	if err != nil {
		return nil, err
	}
	periodStart, _ := parseDate(s.CurrentPeriodStartDate)
	periodEnd, _ := parseDate(s.CurrentPeriodEndDate)
	fyStart, _ := parseDate(s.CurrentFiscalYearStartDate)
	fyEnd, _ := parseDate(s.CurrentFiscalYearEndDate)

	return &domain.FinancialStatement{
		Code:                    s.LocalCode,
		DisclosedDate:           disclosed,
		TypeOfDocument:          s.TypeOfDocument,
		PeriodStart:             periodStart,
		PeriodEnd:               periodEnd,
		FiscalYearStart:         fyStart,
		FiscalYearEnd:           fyEnd,
		NetSales:                s.NetSales,
		OperatingProfit:         s.OperatingProfit,
		OrdinaryProfit:          s.OrdinaryProfit,
		Profit:                  s.Profit,
		EPS:                     s.EarningsPerShare,
		BPS:                     s.BookValuePerShare,
		ForecastNetSales:        s.ForecastNetSales,
		ForecastOperatingProfit: s.ForecastOperatingProfit,
		ForecastProfit:          s.ForecastProfit,
		ForecastEPS:             s.ForecastEarningsPerShare,
	}, nil
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

func newJQuantsClient() *jquants.Client {
	key := os.Getenv("JQUANTS_API_KEY")
	if key == "" {
		log.Fatal("JQUANTS_API_KEY が設定されていません")
	}
	return jquants.New(key)
}

func targetDate() string {
	if d := os.Getenv("TARGET_DATE"); d != "" {
		return d
	}
	t := time.Now()
	if t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, -2)
	} else if t.Weekday() == time.Saturday {
		t = t.AddDate(0, 0, -1)
	}
	return t.Format("2006-01-02")
}
