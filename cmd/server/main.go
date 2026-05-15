package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kusakari/itsumo/internal/auth"
	"github.com/kusakari/itsumo/internal/handler"
	"github.com/kusakari/itsumo/internal/store/postgres"
)

func main() {
	ctx := context.Background()

	pgRepo, err := postgres.New(ctx)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pgRepo.Close()

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	secure := strings.HasPrefix(baseURL, "https://")

	authMgr := auth.NewManager([]byte(mustEnv("SESSION_SECRET")), secure)

	if id := os.Getenv("GOOGLE_CLIENT_ID"); id != "" {
		if err := authMgr.SetupGoogle(ctx, id, mustEnv("GOOGLE_CLIENT_SECRET"), baseURL+"/auth/google/callback"); err != nil {
			log.Fatalf("google oidc: %v", err)
		}
	}
	if id := os.Getenv("APPLE_CLIENT_ID"); id != "" {
		if err := authMgr.SetupApple(ctx, id, mustEnv("APPLE_TEAM_ID"), mustEnv("APPLE_KEY_ID"), mustEnv("APPLE_PRIVATE_KEY"), baseURL+"/auth/apple/callback"); err != nil {
			log.Fatalf("apple oidc: %v", err)
		}
	}

	mux := http.NewServeMux()
	authH := handler.NewAuthHandler(authMgr)
	authH.RegisterRoutes(mux)

	holdingsH := handler.NewHoldingsHandler(pgRepo, authMgr)
	holdingsH.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("itsumo listening on :%s", port)
	if err := http.ListenAndServe(":"+port, authH.Middleware(mux)); err != nil {
		log.Fatal(err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return v
}
