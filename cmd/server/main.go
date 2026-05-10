package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/kusakari/itsumo/internal/domain"
	"github.com/kusakari/itsumo/internal/handler"
	fsrepo "github.com/kusakari/itsumo/internal/store/firestore"
	"github.com/kusakari/itsumo/internal/store/postgres"
)

func main() {
	ctx := context.Background()

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable is required")
	}

	fsRepo, err := fsrepo.New(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore: %v", err)
	}
	defer fsRepo.Close()

	pgRepo, err := postgres.New(ctx)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pgRepo.Close()

	mux := http.NewServeMux()
	handler.New(&repoComposite{fsRepo, pgRepo}).Register(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("itsumo listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

// repoComposite は Firestore (trades/events/votes/companies/earnings) と
// PostgreSQL (prices) を束ねて store.Repository を満たす。
type repoComposite struct {
	*fsrepo.Repo
	pg *postgres.Repo
}

func (r *repoComposite) GetLatestPrice(ctx context.Context, code string) (*domain.DailyPrice, error) {
	return r.pg.GetLatestPrice(ctx, code)
}

func (r *repoComposite) ListPricesByCode(ctx context.Context, code string, limit int) ([]*domain.DailyPrice, error) {
	return r.pg.ListPricesByCode(ctx, code, limit)
}
