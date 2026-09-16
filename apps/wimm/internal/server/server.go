// Package server holds wimmd's HTTP wiring: the mux each listener serves and
// the run loop that shuts them down together.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// APIPrefix is the path every public route hangs under. The web app proxies it
// to wimmd so the browser sees a single origin, which is what keeps the
// WebAuthn origin check, the session cookie and CORS trivial (ADR 0016).
const APIPrefix = "/api"

// Listener is one address and the handler that answers on it.
type Listener struct {
	Name    string
	Addr    string
	Handler http.Handler
}

// Run serves every listener until ctx is cancelled, then shuts them all down.
// It returns the first error that is not a clean shutdown.
func Run(ctx context.Context, log *slog.Logger, shutdownGrace time.Duration, listeners ...Listener) error {
	servers := make([]*http.Server, 0, len(listeners))
	errs := make(chan error, len(listeners))

	for _, l := range listeners {
		ln, err := net.Listen("tcp", l.Addr)
		if err != nil {
			shutdown(context.Background(), servers, shutdownGrace)
			return err
		}

		srv := &http.Server{
			Handler:           l.Handler,
			ReadHeaderTimeout: 10 * time.Second,
		}
		servers = append(servers, srv)

		log.InfoContext(ctx, "listening", "listener", l.Name, "addr", ln.Addr().String())

		go func() {
			if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errs <- err
				return
			}
			errs <- nil
		}()
	}

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-errs:
	}

	shutdown(context.Background(), servers, shutdownGrace)
	return runErr
}

func shutdown(ctx context.Context, servers []*http.Server, grace time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, grace)
	defer cancel()

	for _, srv := range servers {
		_ = srv.Shutdown(ctx)
	}
}
