package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mattmohan-flipp/cas/v2/proxy/store"
	"github.com/mattmohan-flipp/cas/v2/urlscheme"
	"github.com/stretchr/testify/assert"
)

var defaultURL = &url.URL{
	Scheme: "http",
	Host:   "example.com",
}

func TestNewProxy(t *testing.T) {
	urlScheme := urlscheme.NewDefaultURLScheme(defaultURL)
	options := &ProxyOptions{
		RequestProxy:     true,
		ProxyCallbackURL: "http://example.com/callback",
		ProxyStore:       store.NewMemoryProxyStore(),
		UrlScheme:        urlScheme,
	}

	proxy := NewProxy(urlScheme, options)
	assert.NotNil(t, proxy)
	assert.True(t, proxy.IsEnabled())
	assert.Equal(t, options.ProxyCallbackURL, proxy.GetProxyCallbackURL().String())
}

func TestNewProxyInvalidUrl(t *testing.T) {
	urlScheme := urlscheme.NewDefaultURLScheme(defaultURL)
	options := &ProxyOptions{
		RequestProxy:     true,
		ProxyCallbackURL: "\t",
		UrlScheme:        urlScheme,
	}

	proxy := NewProxy(urlScheme, options)
	assert.Nil(t, proxy)
}

func TestNewProxyDisabled(t *testing.T) {
	urlScheme := urlscheme.NewDefaultURLScheme(defaultURL)
	options := &ProxyOptions{
		RequestProxy:     false,
		ProxyCallbackURL: "http://example.com/callback",
		ProxyStore:       store.NewMemoryProxyStore(),
		Logger:           nil,
		UrlScheme:        urlScheme,
	}

	proxy := NewProxy(urlScheme, options)
	assert.NotNil(t, proxy)
	assert.False(t, proxy.IsEnabled())
	assert.Equal(t, options.ProxyCallbackURL, proxy.GetProxyCallbackURL().String())
}
func TestHandle(t *testing.T) {
	urlScheme := urlscheme.NewDefaultURLScheme(defaultURL)
	options := &ProxyOptions{
		RequestProxy:     true,
		ProxyCallbackURL: "http://example.com/callback",
		Logger:           nil,
		UrlScheme:        urlScheme,
	}

	proxy := NewProxy(urlScheme, options)
	assert.NotNil(t, proxy)

	req, err := http.NewRequest(http.MethodGet, "/?pgtIou=testIou&pgtId=testId", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(proxy.Handle)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	pgtId, ok := proxy.proxyStore.Get("testIou")
	assert.True(t, ok)
	assert.Equal(t, "testId", pgtId)
}

func TestWrongMethod(t *testing.T) {
	urlScheme := urlscheme.NewDefaultURLScheme(defaultURL)
	options := &ProxyOptions{
		RequestProxy:     true,
		ProxyCallbackURL: "http://example.com/callback",
		Logger:           nil,
		UrlScheme:        urlScheme,
	}

	proxy := NewProxy(urlScheme, options)
	assert.NotNil(t, proxy)

	req, err := http.NewRequest(http.MethodPost, "/?pgtIou=testIou&pgtId=testId", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(proxy.Handle)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	pgtId, ok := proxy.proxyStore.Get("testIou")
	assert.False(t, ok)
	assert.Equal(t, "", pgtId)
}

func TestGetProxyURL(t *testing.T) {
	urlScheme := urlscheme.NewDefaultURLScheme(defaultURL)
	options := &ProxyOptions{
		RequestProxy:     true,
		ProxyCallbackURL: "http://example.com/callback",
		ProxyStore:       store.NewMemoryProxyStore(),
		Logger:           nil,
		UrlScheme:        urlScheme,
	}

	proxy := NewProxy(urlScheme, options)
	assert.NotNil(t, proxy)

	proxy.proxyStore.Set("testIou", "testPgt")

	proxyURL, err := proxy.GetProxyURL("http://example.com/service", "testIou")
	assert.NoError(t, err)
	assert.Contains(t, proxyURL, "targetService=http%3A%2F%2Fexample.com%2Fservice")
	assert.Contains(t, proxyURL, "pgt=testPgt")
}

func TestGetProxyTgt(t *testing.T) {
	urlScheme := urlscheme.NewDefaultURLScheme(defaultURL)
	options := &ProxyOptions{
		RequestProxy:     true,
		ProxyCallbackURL: "http://example.com/callback",
		ProxyStore:       store.NewMemoryProxyStore(),
		Logger:           nil,
		UrlScheme:        urlScheme,
	}

	proxy := NewProxy(urlScheme, options)
	assert.NotNil(t, proxy)

	proxy.proxyStore.Set("testIou", "testPgt")
	pgt, ok := proxy.GetProxyTgt("testIou")
	assert.True(t, ok)
	assert.Equal(t, "testPgt", pgt)
}
