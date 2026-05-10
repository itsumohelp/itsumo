package firestore

import (
	"context"
	"fmt"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/kusakari/itsumo/internal/domain"
)

// SaveFinancialStatements は決算情報をバッチ書き込みします。
// ドキュメントID は "{code}_{disclosed_YYYYMMDD}" です（冪等性あり）。
func (r *Repo) SaveFinancialStatements(ctx context.Context, stmts []*domain.FinancialStatement) error {
	now := time.Now()
	const batchSize = 500
	for i := 0; i < len(stmts); i += batchSize {
		end := i + batchSize
		if end > len(stmts) {
			end = len(stmts)
		}
		batch := r.client.Batch()
		for _, s := range stmts[i:end] {
			s.SavedAt = now
			docID := s.Code + "_" + s.DisclosedDate.Format("20060102")
			ref := r.client.Collection("financial_statements").Doc(docID)
			batch.Set(ref, s)
		}
		if _, err := batch.Commit(ctx); err != nil {
			return fmt.Errorf("statements batch commit (offset %d): %w", i, err)
		}
	}
	return nil
}

// ListFinancialStatements は指定銘柄の決算情報を開示日昇順で返します。
func (r *Repo) ListFinancialStatements(ctx context.Context, code string) ([]*domain.FinancialStatement, error) {
	docs, err := r.client.Collection("financial_statements").
		Where("code", "==", code).
		OrderBy("disclosed_date", fs.Asc).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("ListFinancialStatements: %w", err)
	}
	result := make([]*domain.FinancialStatement, 0, len(docs))
	for _, doc := range docs {
		s := &domain.FinancialStatement{}
		if err := doc.DataTo(s); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

// SaveEarningsAnnouncements は決算発表予定をバッチ書き込みします。
// ドキュメントID は "{code}_{date_YYYYMMDD}" です（冪等性あり）。
func (r *Repo) SaveEarningsAnnouncements(ctx context.Context, announcements []*domain.EarningsAnnouncement) error {
	now := time.Now()
	const batchSize = 500
	for i := 0; i < len(announcements); i += batchSize {
		end := i + batchSize
		if end > len(announcements) {
			end = len(announcements)
		}
		batch := r.client.Batch()
		for _, a := range announcements[i:end] {
			a.SavedAt = now
			docID := a.Code + "_" + a.Date.Format("20060102")
			ref := r.client.Collection("earnings_announcements").Doc(docID)
			batch.Set(ref, a)
		}
		if _, err := batch.Commit(ctx); err != nil {
			return fmt.Errorf("announcements batch commit (offset %d): %w", i, err)
		}
	}
	return nil
}

// GetNextEarningsAnnouncement は指定銘柄の直近の決算発表予定を返します。
// データがない場合は nil, nil を返します。
func (r *Repo) GetNextEarningsAnnouncement(ctx context.Context, code string) (*domain.EarningsAnnouncement, error) {
	docs, err := r.client.Collection("earnings_announcements").
		Where("code", "==", code).
		Where("date", ">=", time.Now()).
		OrderBy("date", fs.Asc).
		Limit(1).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("GetNextEarningsAnnouncement: %w", err)
	}
	if len(docs) == 0 {
		return nil, nil
	}
	a := &domain.EarningsAnnouncement{}
	if err := docs[0].DataTo(a); err != nil {
		return nil, err
	}
	return a, nil
}
