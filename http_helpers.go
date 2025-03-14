package cas

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mattmohan-flipp/cas/v2/proxy"
)

type key int

const ( // emulating enums is actually pretty ugly in go.
	clientKey key = iota
	authenticationResponseKey
)

// setClient associates a Client with a http.Request.
func setClient(r *http.Request, c *Client) {
	ctx := context.WithValue(r.Context(), clientKey, c)
	r2 := r.WithContext(ctx)
	*r = *r2
}

// getClient retrieves the Client associated with the http.Request.
func getClient(r *http.Request) *Client {
	if c := r.Context().Value(clientKey); c != nil {
		return c.(*Client)
	}

	return nil // explicitly pass along the nil to caller -- conforms to previous impl
}

// RedirectToLogin allows CAS protected handlers to redirect a request
// to the CAS login page.
func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	c := getClient(r)
	if c == nil {
		err := "cas: redirect to cas failed as no client associated with request"
		http.Error(w, err, http.StatusInternalServerError)
		return
	}

	c.RedirectToLogin(w, r)
}

// RedirectToLogout allows CAS protected handlers to redirect a request
// to the CAS logout page.
func RedirectToLogout(w http.ResponseWriter, r *http.Request) {
	c := getClient(r)
	if c == nil {
		err := "cas: redirect to cas failed as no client associated with request"
		http.Error(w, err, http.StatusInternalServerError)
		return
	}

	c.RedirectToLogout(w, r)
}

// setAuthenticationResponse associates an AuthenticationResponse with
// a http.Request.
func setAuthenticationResponse(r *http.Request, a *AuthenticationResponse) {
	ctx := context.WithValue(r.Context(), authenticationResponseKey, a)
	r2 := r.WithContext(ctx)
	*r = *r2
}

// getAuthenticationResponse retrieves the AuthenticationResponse associated
// with a http.Request.
func getAuthenticationResponse(r *http.Request) *AuthenticationResponse {
	if a := r.Context().Value(authenticationResponseKey); a != nil {
		return a.(*AuthenticationResponse)
	}

	return nil // explicitly pass along the nil to caller -- conforms to previous impl
}

// IsAuthenticated indicates whether the request has been authenticated with CAS.
func IsAuthenticated(r *http.Request) bool {
	if a := getAuthenticationResponse(r); a != nil {
		return true
	}

	return false
}

// Username returns the authenticated users username
func Username(r *http.Request) string {
	if a := getAuthenticationResponse(r); a != nil {
		return a.User
	}

	return ""
}

// Attributes returns the authenticated users attributes.
func Attributes(r *http.Request) UserAttributes {
	if a := getAuthenticationResponse(r); a != nil {
		return a.Attributes
	}

	return nil
}

// AuthenticationDate returns the date and time that authentication was performed.
//
// This may return time.IsZero if Authentication Date information is not included
// in the CAS service validation response. This will be the case for CAS 2.0
// protocol servers.
func AuthenticationDate(r *http.Request) time.Time {
	var t time.Time
	if a := getAuthenticationResponse(r); a != nil {
		t = a.AuthenticationDate
	}

	return t
}

// IsNewLogin indicates whether the CAS service ticket was granted following a
// new authentication.
//
// This may incorrectly return false if Is New Login information is not included
// in the CAS service validation response. This will be the case for CAS 2.0
// protocol servers.
func IsNewLogin(r *http.Request) bool {
	if a := getAuthenticationResponse(r); a != nil {
		return a.IsNewLogin
	}

	return false
}

// IsRememberedLogin indicates whether the CAS service ticket was granted by the
// presence of a long term authentication token.
//
// This may incorrectly return false if Remembered Login information is not included
// in the CAS service validation response. This will be the case for CAS 2.0
// protocol servers.
func IsRememberedLogin(r *http.Request) bool {
	if a := getAuthenticationResponse(r); a != nil {
		return a.IsRememberedLogin
	}

	return false
}

// MemberOf returns the list of groups which the user belongs to.
func MemberOf(r *http.Request) []string {
	if a := getAuthenticationResponse(r); a != nil {
		return a.MemberOf
	}

	return nil
}

var errNoClient = errors.New("cas: no client associated with request")
var errNoAuthenticationResponse = errors.New("cas: no authentication response associated with request")
var errProxyUrlError = errors.New("cas: error getting proxy url")
var errProxyTicketError = errors.New("cas: error getting proxy ticket")

func GetProxyTicket(r *http.Request, targetService *url.URL) (string, error) {
	// Get the client from the request context.
	c := getClient(r)
	if c == nil {
		return "", errNoClient
	}

	// Pull the authentication response from the request context.
	a := getAuthenticationResponse(r)
	if a == nil {
		return "", errNoAuthenticationResponse
	}
	if a.ProxyGrantingTicket == "" {
		return "", errors.New("cas: no proxy granting ticket available in authentication response")
	}

	url, err := c.proxy.GetProxyURL(targetService.String(), a.ProxyGrantingTicket)
	if err != nil {
		return "", err
	}

	proxyReq, err := c.client.Get(url)
	if err != nil {
		return "", err
	}

	response, err := io.ReadAll(proxyReq.Body)
	defer proxyReq.Body.Close()
	if err != nil {
		return "", err
	}

	var proxyResponse proxy.XmlProxyResponse
	xml.Unmarshal(response, &proxyResponse)

	if proxyResponse.Success.ProxyTicket == "" {
		return "", errProxyTicketError
	}

	return proxyResponse.Success.ProxyTicket, nil
}
