package cas

import (
	"log/slog"
	"net/http"
)

// Handler returns a standard http.HandlerFunc, which will check the authenticated status (redirect user go login if needed)
// If the user pass the authenticated check, it will call the h's ServeHTTP method
func (c *Client) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.logger.Info("handling request", slog.String("method", r.Method), slog.String("path", r.URL.String()))

		setClient(r, c)

		if !IsAuthenticated(r) {
			RedirectToLogin(w, r)
			return
		}

		if r.URL.Path == "/logout" {
			RedirectToLogout(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}
