package proxy

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/mattmohan-flipp/cas/v2/proxy/store"
	"github.com/mattmohan-flipp/cas/v2/urlscheme"
)

type Proxy struct {
	requestProxy     bool
	proxyCallbackURL *url.URL
	proxyStore       store.ProxyStore
	logger           *slog.Logger
	urlScheme        urlscheme.URLScheme
}

type ProxyOptions struct {
	RequestProxy     bool
	ProxyCallbackURL string
	ProxyStore       store.ProxyStore
	Logger           *slog.Logger
	UrlScheme        urlscheme.URLScheme
}

func NewProxy(urlScheme urlscheme.URLScheme, options *ProxyOptions) *Proxy {
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}

	url, err := url.Parse(options.ProxyCallbackURL)
	if err != nil {
		logger.Error("Failed to parse proxy callback URL", slog.Any("error", err))
		return nil
	}
	proxyStore := options.ProxyStore
	if proxyStore == nil {
		proxyStore = store.NewMemoryProxyStore()
	}

	return &Proxy{
		requestProxy:     options.RequestProxy,
		proxyCallbackURL: url,
		proxyStore:       proxyStore,
		logger:           logger,
		urlScheme:        urlScheme,
	}
}

/**
 * Handle the proxy callback request
 * NOTE: This should return a 200 status code to the CAS server even if the expected params are missing 🙄
 */
func (p *Proxy) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pgtIou := r.URL.Query().Get("pgtIou")
	pgtId := r.URL.Query().Get("pgtId")
	if pgtId != "" && pgtIou != "" {
		p.proxyStore.Set(r.URL.Query().Get("pgtIou"), r.URL.Query().Get("pgtId"))
	}
	w.WriteHeader(http.StatusOK)
}

func (p Proxy) IsEnabled() bool {
	return p.requestProxy
}

func (p Proxy) GetProxyCallbackURL() *url.URL {
	return p.proxyCallbackURL
}

var errProxyIouNotFound = errors.New("cas: proxy iou not found in store")

func (p Proxy) GetProxyURL(targetService, pgtIou string) (string, error) {
	pgt, ok := p.GetProxyTgt(pgtIou)
	if !ok {
		return "", errProxyIouNotFound
	}
	casUrl, err := p.urlScheme.Proxy()
	if err != nil {
		return "", err
	}
	q := casUrl.Query()
	q.Set("targetService", targetService)
	q.Set("pgt", pgt)
	casUrl.RawQuery = q.Encode()
	return casUrl.String(), nil
}

func (p Proxy) GetProxyTgt(pgtIou string) (string, bool) {
	return p.proxyStore.Get(pgtIou)
}
