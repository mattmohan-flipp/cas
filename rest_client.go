package cas

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/mattmohan-flipp/cas/v2/proxy"
	"github.com/mattmohan-flipp/cas/v2/urlscheme"
)

// https://apereo.github.io/cas/4.2.x/protocol/REST-Protocol.html

// TicketGrantingTicket represents a SSO session for a user, also known as TGT
type TicketGrantingTicket string

// ServiceTicket stands for the access granted by the CAS server to an application for a specific user, also known as ST
type ServiceTicket string

// RestOptions provide options for the RestClient
type RestOptions struct {
	CasURL     *url.URL
	ServiceURL *url.URL
	Client     *http.Client
	URLScheme  urlscheme.URLScheme
	Logger     *slog.Logger
	Proxy      *proxy.Proxy
}

// RestClient uses the rest protocol provided by cas
type RestClient struct {
	urlScheme   urlscheme.URLScheme
	serviceURL  *url.URL
	client      *http.Client
	stValidator *ServiceTicketValidator
	logger      *slog.Logger
	proxy       *proxy.Proxy
}

// NewRestClient creates a new client for the cas rest protocol with the provided options
func NewRestClient(options *RestOptions) *RestClient {
	if options.Logger == nil {
		options.Logger = slog.Default()
	}
	options.Logger.Info("new rest client", slog.Any("options", options))

	var client *http.Client
	if options.Client != nil {
		client = options.Client
	} else {
		client = &http.Client{}
	}

	var urlSchemeInstance urlscheme.URLScheme
	if options.URLScheme != nil {
		urlSchemeInstance = options.URLScheme
	} else {
		urlSchemeInstance = urlscheme.NewDefaultURLScheme(options.CasURL)
	}

	proxyInstance := options.Proxy
	if proxyInstance == nil {
		proxyInstance = proxy.NewProxy(urlSchemeInstance, &proxy.ProxyOptions{RequestProxy: false, Logger: options.Logger})
	}

	return &RestClient{
		urlScheme:   urlSchemeInstance,
		serviceURL:  options.ServiceURL,
		client:      client,
		stValidator: NewServiceTicketValidator(ServiceTicketValidatorOptions{Client: client, CasURL: options.CasURL, Logger: options.Logger}),
		logger:      options.Logger,
		proxy:       proxyInstance,
	}
}

// Handle wraps a http.Handler to provide CAS Rest authentication for the handler.
func (c *RestClient) Handle(h http.Handler) http.Handler {
	return &restClientHandler{
		c: c,
		h: h,
	}
}

// HandleFunc wraps a function to provide CAS Rest authentication for the handler function.
func (c *RestClient) HandleFunc(h func(http.ResponseWriter, *http.Request)) http.Handler {
	return c.Handle(http.HandlerFunc(h))
}

// RequestGrantingTicket returns a new TGT, if the username and password authentication was successful
func (c *RestClient) RequestGrantingTicket(username string, password string) (TicketGrantingTicket, error) {
	// request:
	// POST /cas/v1/tickets HTTP/1.0
	// username=battags&password=password&additionalParam1=paramvalue

	endpoint, err := c.urlScheme.RestGrantingTicket()
	if err != nil {
		return "", err
	}

	values := url.Values{}
	values.Set("username", username)
	values.Set("password", password)

	resp, err := c.client.PostForm(endpoint.String(), values)
	if err != nil {
		return "", err
	}

	// response:
	// 201 Created
	// Location: http://www.whatever.com/cas/v1/tickets/{TGT id}

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("ticket endoint returned status code %v", resp.StatusCode)
	}

	tgt := path.Base(resp.Header.Get("Location"))
	if tgt == "" {
		return "", fmt.Errorf("does not return a valid location header")
	}

	return TicketGrantingTicket(tgt), nil
}

// RequestServiceTicket requests a service ticket with the TGT for the configured service url
func (c *RestClient) RequestServiceTicket(tgt TicketGrantingTicket) (ServiceTicket, error) {
	// request:
	// POST /cas/v1/tickets/{TGT id} HTTP/1.0
	// service={form encoded parameter for the service url}
	endpoint, err := c.urlScheme.RestServiceTicket(string(tgt))
	if err != nil {
		return "", err
	}

	values := url.Values{}
	values.Set("service", c.serviceURL.String())

	resp, err := c.client.PostForm(endpoint.String(), values)
	if err != nil {
		return "", err
	}

	// response:
	// 200 OK
	// ST-1-FFDFHDSJKHSDFJKSDHFJKRUEYREWUIFSD2132

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("service ticket endoint returned status code %v", resp.StatusCode)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return ServiceTicket(data), nil
}

// ValidateServiceTicket validates the service ticket and returns an AuthenticationResponse
func (c *RestClient) ValidateServiceTicket(st ServiceTicket) (*AuthenticationResponse, error) {
	return c.stValidator.ValidateTicket(c.serviceURL, string(st), c.proxy)
}

// Logout destroys the given granting ticket
func (c *RestClient) Logout(tgt TicketGrantingTicket) error {
	// DELETE /cas/v1/tickets/TGT-fdsjfsdfjkalfewrihfdhfaie HTTP/1.0
	endpoint, err := c.urlScheme.RestLogout(string(tgt))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", endpoint.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("could not destroy granting ticket %v, server returned %v", tgt, resp.StatusCode)
	}

	return nil
}
