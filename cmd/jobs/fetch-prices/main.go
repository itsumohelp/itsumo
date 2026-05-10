// fetch-prices は J-Quants API から全上場銘柄の日次終値を取得し、
// PostgreSQL の daily_prices テーブルに保存する Cloud Run Job です。
// Cloud Scheduler から毎営業日 17:30 JST に起動することを想定しています。
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/kusakari/itsumo/internal/domain"
	"github.com/kusakari/itsumo/internal/jquants"
	fsrepo "github.com/kusakari/itsumo/internal/store/firestore"
	"github.com/kusakari/itsumo/internal/store/postgres"
)

func main() {
	ctx := context.Background()

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID is required")
	}

	// TARGET_DATE が指定されていればその日付、なければ直前の営業日を使う
	date := targetDate()
	log.Printf("fetch-prices: date=%s", date)

	jq := newJQuantsClient()

	fsRepo, err := fsrepo.New(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore.New: %v", err)
	}
	defer fsRepo.Close()

	pgRepo, err := postgres.New(ctx)
	if err != nil {
		log.Fatalf("postgres.New: %v", err)
	}
	defer pgRepo.Close()

	log.Printf("fetching daily quotes from J-Quants...")
	quotes, err := jq.DailyQuotes(ctx, date)
	if err != nil {
		log.Fatalf("DailyQuotes: %v", err)
	}
	log.Printf("fetched %d quotes", len(quotes))

	if len(quotes) == 0 {
		log.Println("no quotes returned (market closed?). exiting.")
		return
	}

	nameByCode := map[string]string{}
	industryByCode := map[string]string{}
	sectorByCode := map[string]string{}
	log.Printf("fetching equities master for company names...")
	masters, err := jq.EquitiesMaster(ctx, date)
	if err != nil {
		log.Printf("warn: EquitiesMaster: %v (continue without names)", err)
	} else {
		companies := make([]*domain.Company, 0, len(masters))
		for _, m := range masters {
			if m.Code == "" || m.Name() == "" {
				continue
			}
			nameByCode[m.Code] = m.Name()
			industryByCode[m.Code] = m.S33Nm
			sectorByCode[m.Code] = m.S17Nm
			companies = append(companies, &domain.Company{
				Code:     m.Code,
				Name:     m.Name(),
				Industry: m.S33Nm,
				Sector:   m.S17Nm,
			})
		}
		if err := fsRepo.UpsertCompanies(ctx, companies); err != nil {
			log.Printf("warn: UpsertCompanies: %v", err)
		}
		log.Printf("fetched %d master rows (named codes=%d)", len(masters), len(nameByCode))
	}

	prices := make([]*domain.DailyPrice, 0, len(quotes))
	for _, q := range quotes {
		d, err := parseQuoteDate(q.Date)
		if err != nil {
			log.Printf("skip: invalid date %q: %v", q.Date, err)
			continue
		}
		prices = append(prices, &domain.DailyPrice{
			Code:     q.Code,
			Name:     nameByCode[q.Code],
			Industry: industryByCode[q.Code],
			Sector:   sectorByCode[q.Code],
			Date:     d,
			Open:     q.OpenValue(),
			High:     q.HighValue(),
			Low:      q.LowValue(),
			Close:    q.CloseValue(),
			Volume:   q.VolumeValue(),
		})
	}

	log.Printf("saving %d prices to PostgreSQL...", len(prices))
	if err := pgRepo.SaveDailyPrices(ctx, prices); err != nil {
		log.Fatalf("SaveDailyPrices: %v", err)
	}
	log.Printf("done.")
}

func newJQuantsClient() *jquants.Client {
	key := os.Getenv("JQUANTS_API_KEY")
	if key == "" {
		log.Fatal("JQUANTS_API_KEY が設定されていません")
	}
	return jquants.New(key)
}

// targetDate は TARGET_DATE 環境変数か、直前の平日を返します（YYYY-MM-DD 形式）。
func targetDate() string {
	if d := os.Getenv("TARGET_DATE"); d != "" {
		return d
	}
	// 月曜なら金曜（-3日）、それ以外は前日
	t := time.Now().AddDate(0, 0, -1)
	if t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, -1)
	} else if t.Weekday() == time.Saturday {
		t = t.AddDate(0, 0, -1)
	}
	return t.Format("2006-01-02")
}

func parseQuoteDate(s string) (time.Time, error) {
	if d, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return d, nil
	}
	return time.ParseInLocation("20060102", s, time.Local)
}
