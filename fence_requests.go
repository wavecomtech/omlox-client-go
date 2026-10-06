// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"context"
	"net/http"
)

// FencesAPI is a simple wrapper around the client for fences requests.
type FencesAPI struct {
	client *Client
}

// List lists all fences.
func (c *FencesAPI) List(ctx context.Context) ([]Fence, error) {
	requestPath := "/fences/summary"

	return sendRequestParseResponseList[Fence](
		ctx,
		c.client,
		http.MethodGet,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)
}
