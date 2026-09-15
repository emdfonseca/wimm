package server

import (
	"net"
	"net/http"
)

// PublicMux is the mux the browser reaches, mounted under APIPrefix.
//
// Connect handlers register at "/<package>.<Service>/<Method>", so they are
// stripped of the prefix before dispatch: the web app proxies /api to here and
// the path the client builds is the path Connect expects.
func PublicMux(routes ...Route) http.Handler {
	inner := http.NewServeMux()
	for _, r := range routes {
		inner.Handle(r.Pattern, r.Handler)
	}
	inner.HandleFunc("GET /health", health)

	outer := http.NewServeMux()
	outer.Handle(APIPrefix+"/", http.StripPrefix(APIPrefix, inner))
	return outer
}

// Route is one pattern and the handler that answers it, relative to APIPrefix.
type Route struct {
	Pattern string
	Handler http.Handler
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
}

// LoopbackOnly rejects anything that did not arrive over the loopback
// interface. The operator listener binds to loopback by default; this is the
// second line, for a deployment that binds it wider by mistake.
func LoopbackOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
