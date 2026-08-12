// Command worker runs the scheduled job-ingestion process: it polls
// Arbeitnow and Remotive on their respective cadences and upserts
// normalized jobs into Postgres.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/heitorsfreitass/job-radar/internal/adapters/outbound/arbeitnow"
	"github.com/heitorsfreitass/job-radar/internal/adapters/outbound/postgres"
	rediscache "github.com/heitorsfreitass/job-radar/internal/adapters/outbound/redis"
	"github.com/heitorsfreitass/job-radar/internal/adapters/outbound/remotive"
	"github.com/heitorsfreitass/job-radar/internal/config"
	"github.com/heitorsfreitass/job-radar/internal/scheduler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	connectCtx, cancelConnect := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelConnect()

	pool, err := postgres.Connect(connectCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	cache, err := rediscache.New(connectCtx, cfg.RedisAddr, cfg.RedisDB)
	if err != nil {
		log.Fatalf("connect to redis: %v", err)
	}
	defer cache.Close()

	repo := postgres.NewJobsRepository(pool)
	httpClient := &http.Client{Timeout: 30 * time.Second}

	sched := scheduler.New(
		repo,
		cache,
		arbeitnow.New(httpClient),
		remotive.New(httpClient),
		cfg.RemotiveMaxRequestsPerDay,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("worker starting (remotive daily cap: %d requests)", cfg.RemotiveMaxRequestsPerDay)
	sched.Run(ctx)
	log.Println("worker stopped")
}
