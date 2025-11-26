// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"nhooyr.io/websocket"
)

// Client manages communication with Omlox™ Hub client.
type Client struct {
	mu sync.RWMutex

	// the configuration object is immutable after the client has been initialized
	configuration ClientConfiguration

	baseAddress *url.URL

	client *http.Client

	Trackables TrackablesAPI
	Providers  ProvidersAPI

	// websockets client fields

	lifecycleWg sync.WaitGroup
	cancel      context.CancelFunc

	// websockets connection
	conn   *websocket.Conn
	closed bool

	// reconnection support
	reconnectCtx    context.Context
	reconnectCancel context.CancelFunc
	reconnecting    bool

	// subscriptions
	subs map[int]*Subcription

	// pending subscription awaiting for subscription ID from the server
	// can only be one subscription per client awaiting for subscription.
	pending chan chan struct {
		sid int
		err error
	}
}

// New returns a new client decorated with the given configuration options
func New(addr string, options ...ClientOption) (*Client, error) {
	configuration := DefaultConfiguration()

	for _, opt := range options {
		if opt != nil {
			if err := opt(&configuration); err != nil {
				return nil, err
			}
		}
	}

	return newClient(addr, configuration)
}

// newClient returns a new Omlox™ Hub client with a copy of the given configuration
func newClient(addr string, configuration ClientConfiguration) (*Client, error) {
	address, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	c := Client{
		configuration: configuration,

		// configured or default HTTP client
		client: configuration.HTTPClient,

		baseAddress: address,

		closed: true,
		pending: make(chan chan struct {
			sid int
			err error
		}, 1),
		subs: make(map[int]*Subcription),
	}

	c.Trackables = TrackablesAPI{
		client: &c,
	}

	c.Providers = ProvidersAPI{
		client: &c,
	}

	return &c, nil
}

// sendStructuredRequestParseResponse constructs a structured request, sends it, and parses the response
func sendStructuredRequestParseResponse[ResponseT any](
	ctx context.Context,
	client *Client,
	method string,
	path string,
	body any,
	parameters url.Values,
	headers http.Header,
) (*ResponseT, error) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("could not encode request body: %w", err)
	}

	return sendRequestParseResponse[ResponseT](
		ctx,
		client,
		method,
		path,
		&buf,
		parameters,
		headers,
	)
}

// sendRequestParseResponse constructs a request, sends it, and parses the response.
func sendRequestParseResponse[ResponseT any](
	ctx context.Context,
	client *Client,
	method string,
	path string,
	body io.Reader,
	parameters url.Values,
	headers http.Header,
) (*ResponseT, error) {
	// apply the client-level request timeout, if set
	if client.configuration.RequestTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, client.configuration.RequestTimeout)
		defer cancel()
	}

	// TODO: set User-Agent and Content-Type headers

	req, err := client.newRequest(ctx, method, path, body, parameters, headers)
	if err != nil {
		return nil, err
	}

	resp, err := client.send(ctx, req)
	if err != nil || resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := isResponseError(resp); err != nil {
		return nil, err
	}

	return parseResponse[ResponseT](resp.Body)
}

// sendRequestParseResponse constructs a request, sends it, and parses the response.
func sendRequestParseResponseList[ResponseT any](
	ctx context.Context,
	client *Client,
	method string,
	path string,
	body io.Reader,
	parameters url.Values,
	headers http.Header,
) ([]ResponseT, error) {
	// apply the client-level request timeout, if set
	if client.configuration.RequestTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, client.configuration.RequestTimeout)
		defer cancel()
	}

	// TODO: set User-Agent and Content-Type headers

	req, err := client.newRequest(ctx, method, path, body, parameters, headers)
	if err != nil {
		return nil, err
	}

	resp, err := client.send(ctx, req)
	if err != nil || resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := isResponseError(resp); err != nil {
		return nil, err
	}

	return parseResponseList[ResponseT](resp.Body)
}

// newRequest constructs a new request.
func (c *Client) newRequest(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	parameters url.Values,
	headers http.Header,
) (*http.Request, error) {
	// concatenate the base address with the given path
	url := c.baseAddress.JoinPath(path)

	// add query parameters (if any)
	if len(parameters) != 0 {
		url.RawQuery = parameters.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		return nil, fmt.Errorf("could not create '%s %s' request: %w", method, url.String(), err)
	}

	// populate request headers
	if headers != nil {
		req.Header = headers
	}

	return req, nil
}

// send sends the given request to Omlox.
func (c *Client) send(ctx context.Context, req *http.Request) (*http.Response, error) {
	// block on the rate limiter, if set
	if c.configuration.RateLimiter != nil {
		c.configuration.RateLimiter.Wait(ctx)
	}

	return c.client.Do(req)
}

// parseResponse fully consumes the given response body without closing it and
// parses the data into a generic Response[T] structure. If the response body
// is empty, a nil value will be returned.
func parseResponse[T any](responseBody io.Reader) (*T, error) {
	// First, read the data into a buffer. This is not super efficient but we
	// want to know if we actually have a body or not.
	var buf bytes.Buffer

	_, err := buf.ReadFrom(responseBody)
	if err != nil {
		return nil, err
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	var response T
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// parseResponseList fully consumes the given response body without closing it and
// parses the data into a generic T structure list. If the response body
// is empty, a empty T list will be returned.
func parseResponseList[T any](responseBody io.Reader) ([]T, error) {
	// First, read the data into a buffer. This is not super efficient but we
	// want to know if we actually have a body or not.
	var buf bytes.Buffer

	_, err := buf.ReadFrom(responseBody)
	if err != nil {
		return nil, err
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	var response []T
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		return nil, err
	}

	return response, nil
}
