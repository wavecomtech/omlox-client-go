// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"context"
	"net/http"
)

// ProvidersAPI is a simple wrapper around the client for location provider requests.
type ProvidersAPI struct {
	client *Client
}

// List lists all location providers.
func (c *ProvidersAPI) List(ctx context.Context) ([]LocationProvider, error) {
	requestPath := "/providers/summary"

	return sendRequestParseResponseList[LocationProvider](
		ctx,
		c.client,
		http.MethodGet,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)
}

// IDs lists all location providers IDs.
func (c *ProvidersAPI) IDs(ctx context.Context) ([]string, error) {
	requestPath := "/providers"

	return sendRequestParseResponseList[string](
		ctx,
		c.client,
		http.MethodGet,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)
}

// Create creates a location provider.
func (c *ProvidersAPI) Create(ctx context.Context, provider LocationProvider) (*LocationProvider, error) {
	requestPath := "/providers"

	return sendStructuredRequestParseResponse[LocationProvider](
		ctx,
		c.client,
		http.MethodPost,
		requestPath,
		provider,
		nil, // request query parameters
		nil, // request headers
	)
}

// DeleteAll deletes all location providers.
func (c *ProvidersAPI) DeleteAll(ctx context.Context) error {
	requestPath := "/providers"

	_, err := sendRequestParseResponse[struct{}](
		ctx,
		c.client,
		http.MethodDelete,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)

	return err
}

// Get gets a location providers.
func (c *ProvidersAPI) Get(ctx context.Context, id string) (*LocationProvider, error) {
	requestPath := "/providers/" + id

	return sendRequestParseResponse[LocationProvider](
		ctx,
		c.client,
		http.MethodGet,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)
}

// Delete deletes a location provider.
func (c *ProvidersAPI) Delete(ctx context.Context, id string) error {
	requestPath := "/providers/" + id

	_, err := sendRequestParseResponse[struct{}](
		ctx,
		c.client,
		http.MethodDelete,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)

	return err
}

// UpdateLocation updates the location of a location provider.
func (c *ProvidersAPI) UpdateLocation(ctx context.Context, location Location, id string) error {
	requestPath := "/providers/" + id + "/location"

	_, err := sendStructuredRequestParseResponse[struct{}](
		ctx,
		c.client,
		http.MethodPut,
		requestPath,
		location,
		nil, // request query parameters
		nil, // request headers
	)

	return err
}
