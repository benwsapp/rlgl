package server

import (
	"log/slog"
	"net/http"
)

func CSRFMiddleware(next http.Handler, trustedOrigins ...string) http.Handler {
	cop := http.NewCrossOriginProtection()

	for _, origin := range trustedOrigins {
		err := cop.AddTrustedOrigin(origin)
		if err != nil {
			slog.Warn("failed to add trusted origin", "origin", origin, "error", err)
		}
	}

	cop.SetDenyHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {
		slog.Warn("CSRF check failed",
			"method", req.Method,
			"path", req.URL.Path,
			"remote_addr", req.RemoteAddr,
			"origin", req.Header.Get("Origin"),
			"referer", req.Header.Get("Referer"),
			"sec_fetch_site", req.Header.Get("Sec-Fetch-Site"),
		)
		http.Error(responseWriter, "CSRF check failed", http.StatusBadRequest)
	}))

	return SecurityHeaders(cop.Handler(next))
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {
		responseWriter.Header().Set("X-Content-Type-Options", "nosniff")

		responseWriter.Header().Set("X-Frame-Options", "DENY")

		responseWriter.Header().Set("X-XSS-Protection", "1; mode=block")

		responseWriter.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'"
		responseWriter.Header().Set("Content-Security-Policy", csp)

		next.ServeHTTP(responseWriter, req)
	})
}
