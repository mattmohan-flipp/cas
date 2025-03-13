package cas

import (
	"log/slog"
	"net/http"
)

const (
	sessionCookieName = "_cas_session"
)

// clientHandler handles CAS Protocol HTTP requests
type clientHandler struct {
	c *Client
	h http.Handler
}

// ServeHTTP handles HTTP requests, processes CAS requests
// and passes requests up to its child http.Handler.
func (ch *clientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch.c.logger.Info("handling request", slog.String("method", r.Method), slog.String("path", r.URL.String()))

	setClient(r, ch.c)

	if isSingleLogoutRequest(r) {
		ch.performSingleLogout(w, r)
		return
	}

	ch.c.getSession(w, r)
	ch.h.ServeHTTP(w, r)
	return
}

// isSingleLogoutRequest determines if the http.Request is a CAS Single Logout Request.
//
// The rules for a SLO request are, HTTP POST urlencoded form with a logoutRequest parameter.
func isSingleLogoutRequest(r *http.Request) bool {
	if r.Method != "POST" {
		return false
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/x-www-form-urlencoded" {
		return false
	}

	if v := r.FormValue("logoutRequest"); v == "" {
		return false
	}

	return true
}

// performSingleLogout processes a single logout request
func (ch *clientHandler) performSingleLogout(w http.ResponseWriter, r *http.Request) {
	rawXML := r.FormValue("logoutRequest")
	logoutRequest, err := parseLogoutRequest([]byte(rawXML))

	if err != nil {
		ch.c.logger.Error("error parsing logout request", slog.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := ch.c.tickets.Delete(logoutRequest.SessionIndex); err != nil {
		ch.c.logger.Error("error removing ticket", slog.String("err", err.Error()))

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ch.c.deleteSession(logoutRequest.SessionIndex)

	w.WriteHeader(http.StatusOK)
}
