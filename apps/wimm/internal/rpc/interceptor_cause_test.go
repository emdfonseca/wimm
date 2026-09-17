package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

// End to end through the interceptor: a handler returning an unmapped error
// must produce a log line naming the cause.
func TestTheInterceptorLogsTheCauseOfAnInternalError(t *testing.T) {
	var out bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&out, nil))

	handler := AccessLog(log)(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		return nil, toConnectError(errors.New("dial tcp 10.0.0.7:5432: connection refused"))
	})

	_, err := handler(context.Background(), connect.NewRequest(&struct{}{}))
	if err == nil {
		t.Fatal("no error")
	}

	var line map[string]any
	if decodeErr := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &line); decodeErr != nil {
		t.Fatalf("log line is not JSON: %v (%s)", decodeErr, out.String())
	}

	cause, ok := line["cause"].(string)
	if !ok {
		t.Fatalf("no cause in the log line: %s", out.String())
	}
	if !strings.Contains(cause, "connection refused") {
		t.Errorf("cause = %q, want the original failure", cause)
	}
	if code := line["code"]; code != "internal" {
		t.Errorf("code = %v", code)
	}
}

// A mapped refusal names its condition too. Codes collide by design — consent
// having run out and a bank exposing no accounts are both FailedPrecondition —
// so the code alone cannot say which happened, which is exactly what a real
// PayPal refusal looked like in a log.
func TestTheInterceptorLogsTheCauseOfAMappedRefusal(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"no accounts", banking.ErrNoAccounts, "exposed no accounts"},
		{"consent expired", banking.ErrConsentExpired, "access to the bank has run out"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			log := slog.New(slog.NewJSONHandler(&out, nil))

			handler := AccessLog(log)(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
				return nil, toConnectError(fmt.Errorf("completing a connection: %w", tc.err))
			})

			if _, err := handler(context.Background(), connect.NewRequest(&struct{}{})); err == nil {
				t.Fatal("no error")
			}

			var line map[string]any
			if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &line); err != nil {
				t.Fatalf("log line is not JSON: %v", err)
			}

			cause, ok := line["cause"].(string)
			if !ok {
				t.Fatalf("no cause in the log line: %s", out.String())
			}
			if !strings.Contains(cause, tc.want) {
				t.Errorf("cause = %q, want it to mention %q", cause, tc.want)
			}
			// Both are the same code, which is the point.
			if line["code"] != "failed_precondition" {
				t.Errorf("code = %v", line["code"])
			}
		})
	}
}
