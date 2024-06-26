// Copyright (c) 2019-2023, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-log/log"
)

// Config contains the client configuration.
type Config struct {
	// Base URL of the service.
	BaseURL string
	// Auth token to include in the Authorization header of each request (if supplied).
	AuthToken string
	// User agent to include in each request (if supplied).
	UserAgent string
	// HTTPClient to use to make HTTP requests (if supplied).
	HTTPClient *http.Client
	// Logger to be used when output is generated
	Logger log.Logger
}

// DefaultConfig is a configuration that uses default values.
var DefaultConfig = &Config{}

// Client describes the client details.
type Client struct {
	baseURL    *url.URL
	authToken  string
	userAgent  string
	httpClient *http.Client
	logger     log.Logger
}

const defaultBaseURL = ""

// NewClient sets up a new Cloud-Library Service client with the specified base URL and auth token.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig
	}

	// Determine base URL
	bu := defaultBaseURL
	if cfg.BaseURL != "" {
		bu = cfg.BaseURL
	}

	if bu == "" {
		return nil, fmt.Errorf("no BaseURL supplied")
	}

	// If baseURL has a path component, ensure it is terminated with a separator, to prevent
	// url.ResolveReference from stripping the final component of the path when constructing
	// request URL.
	if !strings.HasSuffix(bu, "/") {
		bu += "/"
	}

	baseURL, err := url.Parse(bu)
	if err != nil {
		return nil, err
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported protocol scheme %q", baseURL.Scheme)
	}

	c := &Client{
		baseURL:   baseURL,
		authToken: cfg.AuthToken,
		userAgent: cfg.UserAgent,
	}

	// Set HTTP client
	if cfg.HTTPClient != nil {
		c.httpClient = cfg.HTTPClient
	} else {
		c.httpClient = http.DefaultClient
	}

	if cfg.Logger != nil {
		c.logger = cfg.Logger
	} else {
		c.logger = log.DefaultLogger
	}

	return c, nil
}

// newRequest returns a new Request given a method, relative path, rawQuery, and (optional) body.
func (c *Client) newRequest(ctx context.Context, method, path, rawQuery string, body io.Reader) (*http.Request, error) {
	u := c.baseURL.ResolveReference(&url.URL{
		Path:     path,
		RawQuery: rawQuery,
	})

	r, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if v := c.authToken; v != "" {
		if err := (bearerTokenCredentials{authToken: v}).ModifyRequest(r); err != nil {
			return nil, err
		}
	}

	if v := c.userAgent; v != "" {
		r.Header.Set("User-Agent", v)
	}

	return r, nil
}
