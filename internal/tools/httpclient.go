package tools

import (
	"crypto/tls"
	"net/http"
	"time"
)

const (
	httpClientMaxIdleConns        = 20
	httpClientMaxConnsPerHost     = 20
	httpClientMaxIdleConnsPerHost = 20
	httpClientTimeout             = 60 * time.Second
)

func init() {
	InitialCerts()
}

func NewHttpsClient(domain string, options ...Option) *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = httpClientMaxIdleConns
	t.MaxConnsPerHost = httpClientMaxConnsPerHost
	t.MaxIdleConnsPerHost = httpClientMaxIdleConnsPerHost

	t.TLSClientConfig = &tls.Config{
		RootCAs:    SystemCertPool(),
		ServerName: domain,
	}

	client := &http.Client{
		Timeout:   httpClientTimeout,
		Transport: t,
	}

	for _, option := range options {
		option(client)
	}
	return client
}

type Option func(c *http.Client)

func WithMaxIdleConn(maxIdleConns int) Option {
	return func(c *http.Client) {
		c.Transport.(*http.Transport).MaxIdleConns = maxIdleConns
	}
}

func WithMaxConnsPerHost(maxConnsPerHost int) Option {
	return func(c *http.Client) {
		c.Transport.(*http.Transport).MaxConnsPerHost = maxConnsPerHost
	}
}

func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) Option {
	return func(c *http.Client) {
		c.Transport.(*http.Transport).MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *http.Client) {
		c.Timeout = timeout
	}
}
