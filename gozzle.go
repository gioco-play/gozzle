package gozzle

import (
	"fmt"
	"net/http"
	"net/url"
)

var UserAgentDefault = fmt.Sprintf("Gozzle/%d", 1)

func New(method string, u *url.URL) *Request {
	header := http.Header{}
	header.Add("User-Agent", UserAgentDefault)

	return &Request{
		method: method,
		url:      u,
		header: header,
	}
}

// Post creates POST request
func Post(u *url.URL) *Request {
	return New(http.MethodPost, u)
}

// Get creates GET request
func Get(u *url.URL) *Request {
	return New(http.MethodGet, u)
}

// Put creates PUT request
func Put(url *url.URL) *Request {
	return New(http.MethodPut, url)
}

// Delete creates DELETE request
func Delete(url *url.URL) *Request {
	return New(http.MethodDelete, url)
}

func getHeaders(headers http.Header) map[string]string {
	h := map[string]string{}

	for k := range headers {
		h[k] = headers.Get(k)
	}

	return h
}