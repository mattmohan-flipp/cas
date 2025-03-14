package cas

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/mattmohan-flipp/cas/v2/proxy"
)

type ServiceTicketValidatorOptions struct {
	Client *http.Client
	CasURL *url.URL
	Logger *slog.Logger
}

// NewServiceTicketValidator create a new *ServiceTicketValidator
func NewServiceTicketValidator(options ServiceTicketValidatorOptions) *ServiceTicketValidator {
	return &ServiceTicketValidator{
		client: options.Client,
		casURL: options.CasURL,
		logger: options.Logger,
	}
}

// ServiceTicketValidator is responsible for the validation of a service ticket
type ServiceTicketValidator struct {
	client *http.Client
	casURL *url.URL
	logger *slog.Logger
}

// ValidateTicket validates the service ticket for the given server. The method will try to use the service validate
// endpoint of the cas >= 2 protocol, if the service validate endpoint not available, the function will use the cas 1
// validate endpoint.
func (validator *ServiceTicketValidator) ValidateTicket(serviceURL *url.URL, ticket string, proxy *proxy.Proxy) (*AuthenticationResponse, error) {
	validator.logger.Info("Validating ticket", slog.String("ticket", ticket), slog.String("serviceURL", serviceURL.String()))

	u, err := validator.ServiceValidateUrl(serviceURL, ticket, proxy)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Add("User-Agent", "Golang CAS client github.com/mattmohan-flipp/cas/v2")

	validator.logger.Info("Attempting ticket validation", slog.String("url", r.URL.String()))

	resp, err := validator.client.Do(r)
	if err != nil {
		return nil, err
	}

	validator.logger.Debug("Request returned", slog.String("status", resp.Status), slog.String("url", r.URL.String()), slog.String("method", r.Method))

	if resp.StatusCode == http.StatusNotFound {
		return validator.validateTicketCas1(serviceURL, ticket)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cas: validate ticket: %v", string(body))
	}

	validator.logger.Debug("Received authentication response", slog.String("response", string(body)))

	success, err := ParseServiceResponse(body)
	if err != nil {
		return nil, err
	}

	validator.logger.Debug("Parsed ServiceResponse", slog.Any("response", success))

	return success, nil
}

// ServiceValidateUrl creates the service validation url for the cas >= 2 protocol.
// TODO the function is only exposed, because of the clients ServiceValidateUrl function
func (validator *ServiceTicketValidator) ServiceValidateUrl(serviceURL *url.URL, ticket string, proxy *proxy.Proxy) (string, error) {
	u, err := validator.casURL.Parse(path.Join(validator.casURL.Path, "serviceValidate"))
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("service", sanitisedURLString(serviceURL))
	q.Add("ticket", ticket)
	if proxy.IsEnabled() {
		q.Add("pgtUrl", sanitisedURLString(proxy.GetProxyCallbackURL()))
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (validator *ServiceTicketValidator) validateTicketCas1(serviceURL *url.URL, ticket string) (*AuthenticationResponse, error) {
	u, err := validator.ValidateUrl(serviceURL, ticket)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Add("User-Agent", "Golang CAS client github.com/mattmohan-flipp/cas/v2")
	validator.logger.Debug("Attempting ticket validation", slog.String("url", r.URL.String()))

	resp, err := validator.client.Do(r)
	if err != nil {
		return nil, err
	}
	validator.logger.Debug("Request returned", slog.String("status", resp.Status), slog.String("url", r.URL.String()), slog.String("method", r.Method))

	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	body := string(data)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cas: validate ticket: %v", body)
	}
	validator.logger.Debug("Received authentication response", slog.String("response", body))

	if body == "no\n\n" {
		return nil, nil // not logged in
	}

	success := &AuthenticationResponse{
		User: body[4 : len(body)-1],
	}

	validator.logger.Debug("Parsed ServiceResponse", slog.Any("response", success))

	return success, nil
}

// ValidateUrl creates the validation url for the cas >= 1 protocol.
// TODO the function is only exposed, because of the clients ValidateUrl function
func (validator *ServiceTicketValidator) ValidateUrl(serviceURL *url.URL, ticket string) (string, error) {
	u, err := validator.casURL.Parse(path.Join(validator.casURL.Path, "validate"))
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("service", sanitisedURLString(serviceURL))
	q.Add("ticket", ticket)
	u.RawQuery = q.Encode()

	return u.String(), nil
}
