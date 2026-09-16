// Command wimmd serves wimm's public and operator APIs.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/config"
	"github.com/xuuid/wimm/apps/wimm/internal/identity"
	"github.com/xuuid/wimm/apps/wimm/internal/rpc"
	"github.com/xuuid/wimm/apps/wimm/internal/server"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

func main() {
	os.Exit(run())
}

func run() int {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadFromEnv()
	if err != nil {
		// Refusing to start is the point: an operator surface reachable with
		// nothing guarding it must never be served (ADR 0016).
		log.ErrorContext(ctx, "refusing to start", "error", err)
		return 2
	}

	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.ErrorContext(ctx, "refusing to start", "error", err)
		return 2
	}
	defer db.Close()

	svc, err := identity.New(db, cfg)
	if err != nil {
		log.ErrorContext(ctx, "refusing to start", "error", err)
		return 2
	}

	listeners := []server.Listener{{
		Name:    "public",
		Addr:    cfg.PublicAddr,
		Handler: server.PublicMux(publicRoute(log, svc, cfg.Origins)),
	}}

	if cfg.OperatorEnabled {
		path, handler := rpc.OperatorHandler(log, rpc.NewOperatorServer(svc), cfg.OperatorCredential)
		mux := http.NewServeMux()
		mux.Handle(path, handler)

		listeners = append(listeners, server.Listener{
			Name: "operator",
			Addr: cfg.OperatorAddr,
			// Loopback by default in configuration; enforced here as well, so
			// a deployment that binds it wider does not thereby open it.
			Handler: server.LoopbackOnly(mux),
		})
	}

	log.InfoContext(ctx, "starting",
		"relying_party_id", cfg.RelyingPartyID,
		"origins", cfg.Origins,
		"operator_listener", cfg.OperatorEnabled,
		"sweep_interval", cfg.SweepInterval)

	// One goroutine beside the listeners, sharing their signal context so it
	// stops when they do. Nothing waits for it on the way out: an interrupted
	// sweep leaves rows that the next one removes.
	go identity.NewSweeper(db, log, cfg).Run(ctx)

	if err := server.Run(ctx, log, 10*time.Second, listeners...); err != nil {
		log.ErrorContext(ctx, "serving", "error", err)
		return 1
	}
	return 0
}

func publicRoute(log *slog.Logger, svc *identity.Service, origins []string) server.Route {
	path, handler := rpc.PublicHandler(log, rpc.NewPublicServer(svc, origins))
	return server.Route{Pattern: path, Handler: handler}
}
