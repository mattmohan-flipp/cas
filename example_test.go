package cas

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
)

type myHandler struct{}

var MyHandler = &myHandler{}
var casURL string

func init() {
	flag.StringVar(&casURL, "url", "", "CAS server URL")
}

func main() {
	Example()
}

func Example() {
	flag.Parse()

	if casURL == "" {
		flag.Usage()
		return
	}

	slog.Info("Starting up")

	m := http.NewServeMux()
	m.Handle("/", MyHandler)

	url, _ := url.Parse(casURL)
	client := NewClient(&Options{
		URL: url,
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: client.Handle(m),
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error from HTTP Server: %v", slog.Any("error", err))
	}

	slog.Info("Shutting down")
}

type templateBinding struct {
	Username   string
	Attributes UserAttributes
}

func (h *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(r) {
		RedirectToLogin(w, r)
		return
	}

	if r.URL.Path == "/logout" {
		RedirectToLogout(w, r)
		return
	}

	w.Header().Add("Content-Type", "text/html")

	tmpl, err := template.New("index.html").Parse(indexHTML)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, error500, err)
		return
	}

	binding := &templateBinding{
		Username:   Username(r),
		Attributes: Attributes(r),
	}

	html := new(bytes.Buffer)
	if err := tmpl.Execute(html, binding); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, error500, err)
		return
	}

	html.WriteTo(w)
}

const indexHTML = `<!DOCTYPE html>
<html>
  <head>
    <title>Welcome {{.Username}}</title>
  </head>
  <body>
    <h1>Welcome {{.Username}} <a href="/logout">Logout</a></h1>
    <p>Your attributes are:</p>
    <ul>{{range $key, $values := .Attributes}}
      <li>{{$len := len $values}}{{$key}}:{{if gt $len 1}}
        <ul>{{range $values}}
          <li>{{.}}</li>{{end}}
        </ul>
      {{else}} {{index $values 0}}{{end}}</li>{{end}}
    </ul>
  </body>
</html>
`

const error500 = `<!DOCTYPE html>
<html>
  <head>
    <title>Error 500</title>
  </head>
  <body>
    <h1>Error 500</h1>
    <p>%v</p>
  </body>
</html>
`
