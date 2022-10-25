package app

import (
	// import built-in packages
	"net/http"
	"net/url"
	"time"
)

// Default configuration values
const (
	DefaultMaxBodySize int64 = 1024 * 1024
	DefaultMaxDuration       = 10 * time.Second
	DefaultHostname          = "httpbin"
)

const (
	jsonContentType = "application/json; encoding=utf-8"
	htmlContentType = "text/html; charset=utf-8"
)

type headersResponse struct {
	Headers http.Header `json:"headers"`
}

type ipResponse struct {
	Origin string `json:"origin"`
}

type userAgentResponse struct {
	UserAgent string `json:"user-agent"`
}

type getResponse struct {
	Args    url.Values  `json:"args"`
	Headers http.Header `json:"headers"`
	Origin  string      `json:"origin"`
	URL     string      `json:"url"`
}

// A generic response for any incoming request that might contain a body
type bodyResponse struct {
	Args    url.Values  `json:"args"`
	Headers http.Header `json:"headers"`
	Origin  string      `json:"origin"`
	URL     string      `json:"url"`

	Data  string              `json:"data"`
	Files map[string][]string `json:"files"`
	Form  map[string][]string `json:"form"`
	JSON  interface{}         `json:"json"`
}

type cookiesResponse map[string]string

type authResponse struct {
	Authorized bool   `json:"authorized"`
	User       string `json:"user"`
}

type gzipResponse struct {
	Headers http.Header `json:"headers"`
	Origin  string      `json:"origin"`
	Gzipped bool        `json:"gzipped"`
}

type deflateResponse struct {
	Headers  http.Header `json:"headers"`
	Origin   string      `json:"origin"`
	Deflated bool        `json:"deflated"`
}

// An actual stream response body will be made up of one or more of these
// structs, encoded as JSON and separated by newlines
type streamResponse struct {
	ID      int         `json:"id"`
	Args    url.Values  `json:"args"`
	Headers http.Header `json:"headers"`
	Origin  string      `json:"origin"`
	URL     string      `json:"url"`
}

type uuidResponse struct {
	UUID string `json:"uuid"`
}

type bearerResponse struct {
	Authenticated bool   `json:"authenticated"`
	Token         string `json:"token"`
}

type hostnameResponse struct {
	Hostname string `json:"hostname"`
}

// HTTPBin contains the business logic
type HTTPBin struct {
	// Max size of an incoming request generated response body, in bytes
	MaxBodySize int64

	// Max duration of a request, for those requests that allow user control
	// over timing (e.g. /delay)
	MaxDuration time.Duration

	// Observer called with the result of each handled request
	Observer Observer

	// Default parameter values
	DefaultParams DefaultParams

	// The hostname to expose via /hostname.
	hostname string
}

// DefaultParams defines default parameter values
type DefaultParams struct {
	DripDuration time.Duration
	DripDelay    time.Duration
	DripNumBytes int64
}

// DefaultDefaultParams defines the DefaultParams that are used by default. In
// general, these should match the original httpbin.org's defaults.
var DefaultDefaultParams = DefaultParams{
	DripDuration: 2 * time.Second,
	DripDelay:    2 * time.Second,
	DripNumBytes: 10,
}

// Handler returns an http.Handler that exposes all HTTPBin endpoints
func (h *HTTPBin) Handler() http.Handler {
	engine := http.NewServeMux()

	engine.HandleFunc("/", methods(h.Index, "GET"))
	engine.HandleFunc("/forms/post", methods(h.FormsPost, "GET"))
	engine.HandleFunc("/encoding/utf8", methods(h.UTF8, "GET"))

	engine.HandleFunc("/delete", methods(h.RequestWithBody, "DELETE"))
	engine.HandleFunc("/get", methods(h.Get, "GET"))
	engine.HandleFunc("/head", methods(h.Get, "HEAD"))
	engine.HandleFunc("/patch", methods(h.RequestWithBody, "PATCH"))
	engine.HandleFunc("/post", methods(h.RequestWithBody, "POST"))
	engine.HandleFunc("/put", methods(h.RequestWithBody, "PUT"))

	engine.HandleFunc("/ip", h.IP)
	engine.HandleFunc("/user-agent", h.UserAgent)
	engine.HandleFunc("/headers", h.Headers)
	engine.HandleFunc("/response-headers", h.ResponseHeaders)
	engine.HandleFunc("/hostname", h.Hostname)

	engine.HandleFunc("/status/", h.Status)
	engine.HandleFunc("/unstable", h.Unstable)

	engine.HandleFunc("/redirect/", h.Redirect)
	engine.HandleFunc("/relative-redirect/", h.RelativeRedirect)
	engine.HandleFunc("/absolute-redirect/", h.AbsoluteRedirect)
	engine.HandleFunc("/redirect-to", h.RedirectTo)

	engine.HandleFunc("/anything/", h.RequestWithBody)
	engine.HandleFunc("/anything", h.RequestWithBody)

	engine.HandleFunc("/cookies", h.Cookies)
	engine.HandleFunc("/cookies/set", h.SetCookies)
	engine.HandleFunc("/cookies/delete", h.DeleteCookies)

	engine.HandleFunc("/basic-auth/", h.BasicAuth)
	engine.HandleFunc("/hidden-basic-auth/", h.HiddenBasicAuth)
	engine.HandleFunc("/digest-auth/", h.DigestAuth)
	engine.HandleFunc("/bearer", h.Bearer)

	engine.HandleFunc("/deflate", h.Deflate)
	engine.HandleFunc("/gzip", h.Gzip)

	engine.HandleFunc("/stream/", h.Stream)
	engine.HandleFunc("/delay/", h.Delay)
	engine.HandleFunc("/drip", h.Drip)

	engine.HandleFunc("/range/", h.Range)
	engine.HandleFunc("/bytes/", h.Bytes)
	engine.HandleFunc("/stream-bytes/", h.StreamBytes)

	engine.HandleFunc("/html", h.HTML)
	engine.HandleFunc("/robots.txt", h.Robots)
	engine.HandleFunc("/deny", h.Deny)

	engine.HandleFunc("/cache", h.Cache)
	engine.HandleFunc("/cache/", h.CacheControl)
	engine.HandleFunc("/etag/", h.ETag)

	engine.HandleFunc("/links/", h.Links)

	engine.HandleFunc("/image", h.ImageAccept)
	engine.HandleFunc("/image/", h.Image)
	engine.HandleFunc("/xml", h.XML)
	engine.HandleFunc("/json", h.JSON)

	engine.HandleFunc("/uuid", h.UUID)
	engine.HandleFunc("/base64/", h.Base64)

	// existing httpbin endpoints that we do not support
	engine.HandleFunc("/brotli", notImplementedHandler)

	// Make sure our ServeMux doesn't "helpfully" redirect these invalid
	// endpoints by adding a trailing slash. See the ServeMux docs for more
	// info: https://golang.org/pkg/net/http/#ServeMux
	engine.HandleFunc("/absolute-redirect", http.NotFound)
	engine.HandleFunc("/basic-auth", http.NotFound)
	engine.HandleFunc("/delay", http.NotFound)
	engine.HandleFunc("/digest-auth", http.NotFound)
	engine.HandleFunc("/hidden-basic-auth", http.NotFound)
	engine.HandleFunc("/redirect", http.NotFound)
	engine.HandleFunc("/relative-redirect", http.NotFound)
	engine.HandleFunc("/status", http.NotFound)
	engine.HandleFunc("/stream", http.NotFound)
	engine.HandleFunc("/bytes", http.NotFound)
	engine.HandleFunc("/stream-bytes", http.NotFound)
	engine.HandleFunc("/links", http.NotFound)

	// Apply global middleware
	var handler http.Handler
	handler = engine
	handler = limitRequestSize(h.MaxBodySize, handler)
	handler = preflight(handler)
	handler = autohead(handler)
	if h.Observer != nil {
		handler = observe(h.Observer, handler)
	}

	return handler
}

// New creates a new HTTPBin instance
func New(opts ...OptionFunc) *HTTPBin {
	h := &HTTPBin{
		MaxBodySize:   DefaultMaxBodySize,
		MaxDuration:   DefaultMaxDuration,
		DefaultParams: DefaultDefaultParams,
		hostname:      DefaultHostname,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// OptionFunc uses the "functional options" pattern to customize an HTTPBin
// instance
type OptionFunc func(*HTTPBin)

// WithDefaultParams sets the default params handlers will use
func WithDefaultParams(defaultParams DefaultParams) OptionFunc {
	return func(h *HTTPBin) {
		h.DefaultParams = defaultParams
	}
}

// WithMaxBodySize sets the maximum amount of memory
func WithMaxBodySize(m int64) OptionFunc {
	return func(h *HTTPBin) {
		h.MaxBodySize = m
	}
}

// WithMaxDuration sets the maximum amount of time httpbin may take to respond
func WithMaxDuration(d time.Duration) OptionFunc {
	return func(h *HTTPBin) {
		h.MaxDuration = d
	}
}

// WithHostname sets the hostname to return via the /hostname endpoint.
func WithHostname(s string) OptionFunc {
	return func(h *HTTPBin) {
		h.hostname = s
	}
}

// WithObserver sets the request observer callback
func WithObserver(o Observer) OptionFunc {
	return func(h *HTTPBin) {
		h.Observer = o
	}
}
