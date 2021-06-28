package main

import (
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type RLHTTPClient struct {
	client      *http.Client
	RateLimiter *rate.Limiter
}

func (c *RLHTTPClient) Do(req *http.Request) (*http.Response, error) {
	err := c.RateLimiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *RLHTTPClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func NewRLHTTPClient(rl *rate.Limiter, disableKeepAlives bool) *RLHTTPClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableKeepAlives = disableKeepAlives
	client := http.Client{Transport: transport, Timeout: 500 * time.Millisecond}
	c := &RLHTTPClient{
		client:      &client,
		RateLimiter: rl,
	}
	return c
}
