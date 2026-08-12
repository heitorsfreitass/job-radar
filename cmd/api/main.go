// Command api serves job-radar's REST API.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	httpadapter "github.com/heitorsfreitass/job-radar/internal/adapters/inbound/http"
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

	router := httpadapter.NewRouter()

	addr := ":" + cfg.APIPort
	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
