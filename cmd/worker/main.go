// Command worker runs the scheduled job-ingestion process. As of this
// skeleton stage it only validates configuration and the database
// connection; the actual ingestion scheduler lands in the next stage.
package main

import (
	"context"
	"log"
	"time"

	"github.com/heitorsfreitass/job-radar/internal/adapters/outbound/postgres"
	"github.com/heitorsfreitass/job-radar/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	log.Printf("worker ready (remotive daily cap: %d requests)", cfg.RemotiveMaxRequestsPerDay)
}
