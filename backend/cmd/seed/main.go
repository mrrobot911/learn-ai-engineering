package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrrobot911/learn-ai-engineering/internal/config"
	"github.com/mrrobot911/learn-ai-engineering/internal/migrate"
	"github.com/mrrobot911/learn-ai-engineering/internal/seed"
	seeddata "github.com/mrrobot911/learn-ai-engineering/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required to load seed content")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("Failed to acquire database connection: %v", err)
	}
	if err := migrate.Apply(ctx, conn.Conn(), os.DirFS("migrations")); err != nil {
		conn.Release()
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	conn.Release()

	report, err := seed.Load(ctx, pool, seeddata.Files)
	if err != nil {
		log.Fatalf("Failed to load seed content: %v", err)
	}
	log.Printf("Seed complete: %s", report)
	if report.Created() == 0 {
		log.Print("No seed content created.")
	}
}
