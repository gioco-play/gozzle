package gozzle

import (
	"bytes"
	"encoding/json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/publicsuffix"
)

// Request HTTP-request struct
type Request struct {
	method  string
	url     string
	header  http.Header
	cookies []*http.Cookie
	body    []byte
	debug   DebugHandler
	options
}

type options struct {
	clientTransport http.RoundTripper
	clientTimeout   time.Duration
}

// String returns request body as string
func (r *Request) String() string {
	return string(r.body)
}

// Timeout set request timeout(sec)
func (r *Request) Timeout(i int) *Request {
	r.options.clientTimeout = time.Duration(i) * time.Second
	return r
}

// Transport set request transport
func (r *Request) Transport(t http.RoundTripper) *Request {
	r.options.clientTransport = t
	return r
}

// Header set request header by key,value
func (r *Request) Header(key, value string) *Request {
	r.header.Set(key, value)
	return r
}

// Headers set request header by map[string]string
func (r *Request) Headers(headers map[string]string) *Request {
	for k, v := range headers {
		r.header.Set(k, v)
	}
	return r
}

// UserAgent sets request custom user agent request header
func (r *Request) UserAgent(userAgent string) *Request {
	r.Header("User-Agent", userAgent)
	return r
}

// Referer sets referer header
func (r *Request) Referer(referer string) *Request {
	r.Header("Referer", referer)
	return r
}

// Cookie sets request cookie
func (r *Request) Cookie(cookie *http.Cookie) *Request {
	r.cookies = append(r.cookies, cookie)
	return r
}

// Debug sets request debug handler func
func (r *Request) Debug(h DebugHandler) *Request {
	r.debug = h
	return r
}

// Trace set tracing by jaeger
func (r *Request) Trace(span trace.Span) *Request {
	r.debug = func(response *Response) {
		span.AddEvent("gozzle", trace.WithAttributes(
			attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(response.Request().String()),
			},
			attribute.KeyValue{
				Key:   "response",
				Value: attribute.StringValue(response.String()),
			},
		))
	}
	return r
}

// GetMethod returns request method
func (r *Request) GetMethod() string {
	return r.method
}

// GetURL returns request URL
func (r *Request) GetURL() string {
	return r.url
}

// GetHeaders returns request headers map
func (r *Request) GetHeaders() map[string]string {
	return getHeaders(r.header)
}

// Body sets request body
func (r *Request) Body(body []byte) (*Response, error) {
	r.body = body
	return r.Do()
}

// JSON sets request JSON and returns response
func (r *Request) JSON(v interface{}) (*Response, error) {
	r.Header("Content-Type", "application/json")

	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return r.Body(body)
}

// Form sends encoded form and returns response
func (r *Request) Form(v url.Values) (*Response, error) {
	r.Header("Content-Type", "application/x-www-form-urlencoded")
	return r.Body([]byte(v.Encode()))
}

// Do send request & returns response
func (r *Request) Do() (*Response, error) {
	client, err := r.client()
	if err != nil {
		return nil, err
	}

	var buf io.Reader
	if len(r.body) > 0 {
		buf = bytes.NewBuffer(r.body)
	}

	request, err := http.NewRequest(r.method, r.url, buf)
	if err != nil {
		return nil, err
	}

	for k := range r.header {
		request.Header.Add(k, r.header.Get(k))
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return r.response(response)
}

func (r *Request) client() (*http.Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	if err != nil {
		return nil, err
	}

	u, err := url.Parse(r.url)
	if err != nil {
		return nil, err
	}

	jar.SetCookies(u, r.cookies)
	client := &http.Client{
		Jar: jar,
	}

	if r.clientTimeout != 0 {
		client.Timeout = r.clientTimeout
	}

	if r.clientTransport != nil {
		client.Transport = r.clientTransport
	}

	if len(r.cookies) > 0 {
		client.Jar.SetCookies(u, r.cookies)
	}

	return client, nil
}

func (r *Request) response(response *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("read error", err)
		return nil, err
	}

	res := &Response{
		request: r,
		status:  response.StatusCode,
		headers: response.Header,
		cookies: response.Cookies(),
		body:    body,
	}

	if r.debug != nil {
		r.debug(res)
	}

	return res, nil
}
