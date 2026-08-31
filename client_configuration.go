// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
)

// WSCheckRetry specifies a policy for handling WebSocket connection retries.
// It is called following each connection attempt with the response (if any) and error values.
// If CheckRetry returns false, the Client stops retrying and returns the error to the caller.
// If CheckRetry returns an error, that error value is returned in lieu of the error from the connection attempt.
type WSCheckRetry func(ctx context.Context, attemptNum int, err error) (bool, error)

// WSBackoff specifies a policy for how long to wait between connection retry attempts.
// It is called after a failing connection attempt to determine the amount of time
// that should pass before trying again.
type WSBackoff func(min, max time.Duration, attemptNum int) time.Duration

// DefaultWSRetryPolicy provides a default callback for WebSocket connection retries,
// which will retry on connection errors and WebSocket close errors (except normal closure).
func DefaultWSRetryPolicy(ctx context.Context, attemptNum int, err error) (bool, error) {
	// Do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err == nil {
		return false, nil
	}

	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return false, err
	}

	// A certificate the client will not accept is a configuration problem, not
	// a transient one. Retrying only delays an error that cannot change, and
	// buries the cause under a run of failed attempts; WithRootCAFile or
	// WithInsecureSkipVerify is what resolves it.
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		return false, err
	}

	return true, nil
}

// DefaultWSBackoff provides exponential backoff based on the attempt number and limited
// by the provided minimum and maximum durations.
func DefaultWSBackoff(min, max time.Duration, attemptNum int) time.Duration {
	mult := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(mult)
	if sleep > max {
		sleep = max
	}
	return sleep
}

// GetDefaultOptions returns default configuration options for the client.
func DefaultConfiguration() ClientConfiguration {
	// Use cleanhttp, which has the same default values as net/http client, but
	// does not share state with other clients (see: gh/hashicorp/go-cleanhttp)
	defaultClient := cleanhttp.DefaultPooledClient()

	return ClientConfiguration{
		HTTPClient:     defaultClient,
		RequestTimeout: 60 * time.Second,

		WSMaxRetries:   2,
		WSMinRetryWait: time.Second,
		WSMaxRetryWait: 30 * time.Second,
		WSBackoff:      DefaultWSBackoff,
		WSCheckRetry:   DefaultWSRetryPolicy,
	}
}

// / ClientConfiguration is used to configure the creation of the client.
type ClientConfiguration struct {
	// HTTPClient is the HTTP client to use for all API requests.
	HTTPClient *http.Client

	// RequestTimeout, given a non-negative value, will apply the timeout to
	// each request function unless an earlier deadline is passed to the
	// request function through context.Context.
	//
	// Default: 60s
	RequestTimeout time.Duration

	// RateLimiter controls how frequently requests are allowed to happen.
	// If this pointer is nil, then there will be no limit set. Note that an
	// empty struct rate.Limiter is equivalent to blocking all requests.
	//
	// Default: nil
	RateLimiter *rate.Limiter

	// UserAgent sets a name for the http client User-Agent header.
	UserAgent string

	// WSMaxRetries sets the maximum number of reconnection attempts for WebSocket connections.
	// Set to -1 for unlimited retries, 0 to disable retries.
	//
	// Default: 2
	WSMaxRetries int

	// WSMinRetryWait sets the minimum time to wait before retrying a WebSocket connection.
	//
	// Default: 1s
	WSMinRetryWait time.Duration

	// WSMaxRetryWait sets the maximum time to wait before retrying a WebSocket connection.
	//
	// Default: 30s
	WSMaxRetryWait time.Duration

	// WSBackoff specifies the policy for how long to wait between WebSocket reconnection attempts.
	//
	// Default: DefaultWSBackoff (exponential backoff)
	WSBackoff WSBackoff

	// WSCheckRetry specifies the policy for handling WebSocket reconnection retries.
	//
	// Default: DefaultWSRetryPolicy
	WSCheckRetry WSCheckRetry

	// WSAutoReconnect enables automatic reconnection when WebSocket connection is lost.
	// When true, the client will automatically attempt to reconnect and resubscribe
	// to all active subscriptions.
	//
	// Default: false (for backward compatibility, but recommended to enable)
	WSAutoReconnect bool
}

// ClientOption is a configuration option to initialize a client.
type ClientOption func(*ClientConfiguration) error

// WithHTTPClient sets the HTTP client to use for all API requests.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *ClientConfiguration) error {
		c.HTTPClient = client
		return nil
	}
}

// WithRequestTimeout, given a non-negative value, will apply the timeout to
// each request function unless an earlier deadline is passed to the request
// function through context.Context.
//
// Default: 60s
func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfiguration) error {
		if timeout < 0 {
			return fmt.Errorf("request timeout must not be negative")
		}
		c.RequestTimeout = timeout
		return nil
	}
}

// WithRateLimiter configures how frequently requests are allowed to happen.
// If this pointer is nil, then there will be no limit set. Note that an
// empty struct rate.Limiter is equivalent to blocking all requests.
//
// Default: nil
func WithRateLimiter(limiter *rate.Limiter) ClientOption {
	return func(c *ClientConfiguration) error {
		c.RateLimiter = limiter
		return nil
	}
}

// WithWSAutoReconnect enables or disables automatic WebSocket reconnection.
// When enabled, the client will automatically attempt to reconnect and resubscribe
// to all active subscriptions when the connection is lost.
//
// Default: false (for backward compatibility)
func WithWSAutoReconnect(enabled bool) ClientOption {
	return func(c *ClientConfiguration) error {
		c.WSAutoReconnect = enabled
		return nil
	}
}

// WithWSMaxRetries sets the maximum number of WebSocket reconnection attempts.
// Set to -1 for unlimited retries, 0 to disable retries.
//
// Default: 2
func WithWSMaxRetries(retries int) ClientOption {
	return func(c *ClientConfiguration) error {
		if retries < -1 {
			return fmt.Errorf("retries must not be less than -1")
		}
		c.WSMaxRetries = retries
		return nil
	}
}

// WithWSRetryWait sets the minimum and maximum time to wait between WebSocket reconnection attempts.
//
// Default: min=1s, max=30s
func WithWSRetryWait(min, max time.Duration) ClientOption {
	return func(c *ClientConfiguration) error {
		if min < 0 {
			return fmt.Errorf("min retry wait must not be negative")
		}
		if max < min {
			return fmt.Errorf("max retry wait must be greater than or equal to min retry wait")
		}
		c.WSMinRetryWait = min
		c.WSMaxRetryWait = max
		return nil
	}
}

// WithWSBackoff sets a custom backoff policy for WebSocket reconnection attempts.
//
// Default: DefaultWSBackoff (exponential backoff)
func WithWSBackoff(backoff WSBackoff) ClientOption {
	return func(c *ClientConfiguration) error {
		c.WSBackoff = backoff
		return nil
	}
}

// WithWSCheckRetry sets a custom retry policy for WebSocket reconnection attempts.
//
// Default: DefaultWSRetryPolicy
func WithWSCheckRetry(checkRetry WSCheckRetry) ClientOption {
	return func(c *ClientConfiguration) error {
		c.WSCheckRetry = checkRetry
		return nil
	}
}

// WithConnectionPoolSettings configures HTTP connection pool settings for better
// performance under high load. This adjusts MaxIdleConns, MaxIdleConnsPerHost,
// and MaxConnsPerHost on the transport.
func WithConnectionPoolSettings(maxIdleConns, maxIdleConnsPerHost, maxConnsPerHost int) ClientOption {
	return func(c *ClientConfiguration) error {
		if c.HTTPClient == nil {
			c.HTTPClient = cleanhttp.DefaultPooledClient()
		}

		transport, ok := c.HTTPClient.Transport.(*http.Transport)
		if !ok {
			return fmt.Errorf("HTTPClient transport must be *http.Transport to configure connection pool")
		}

		transport.MaxIdleConns = maxIdleConns
		transport.MaxIdleConnsPerHost = maxIdleConnsPerHost
		transport.MaxConnsPerHost = maxConnsPerHost

		return nil
	}
}

// transport returns the *http.Transport the configured client uses, so TLS
// settings can be applied to it. It creates the default pooled client when
// none was configured yet.
//
// A client whose transport is a wrapper type (a retrying or instrumenting
// RoundTripper, say) cannot be reached this way. Configure TLS on the inner
// *http.Client before wrapping it and pass the wrapper with WithHTTPClient.
func (c *ClientConfiguration) transport() (*http.Transport, error) {
	if c.HTTPClient == nil {
		c.HTTPClient = cleanhttp.DefaultPooledClient()
	}

	transport, ok := c.HTTPClient.Transport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("HTTPClient transport must be *http.Transport to configure TLS")
	}

	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return transport, nil
}

// WithTLSConfig sets the TLS configuration used for both the HTTPS API calls
// and the WSS websockets connection, which share the client's transport.
//
// Because it configures the transport of the client set at the time it runs,
// it must be passed after WithHTTPClient.
func WithTLSConfig(config *tls.Config) ClientOption {
	return func(c *ClientConfiguration) error {
		if config == nil {
			return fmt.Errorf("tls config must not be nil")
		}

		transport, err := c.transport()
		if err != nil {
			return err
		}

		transport.TLSClientConfig = config

		return nil
	}
}

// WithInsecureSkipVerify disables verification of the hub's TLS certificate
// for both HTTPS and WSS.
//
// This is for reaching a hub that serves a self-signed or otherwise
// unverifiable certificate — a development or staging deployment behind a
// reverse proxy with no CA-issued certificate. It disables the protection
// against an impersonated hub, so prefer WithRootCAFile, which keeps
// verification on, wherever the CA certificate can be obtained.
//
// Because it configures the transport of the client set at the time it runs,
// it must be passed after WithHTTPClient.
func WithInsecureSkipVerify(skip bool) ClientOption {
	return func(c *ClientConfiguration) error {
		transport, err := c.transport()
		if err != nil {
			return err
		}

		transport.TLSClientConfig.InsecureSkipVerify = skip

		return nil
	}
}

// WithRootCAFile trusts the PEM-encoded CA certificates in the given file when
// verifying the hub's certificate, in addition to the system roots.
//
// This is the verifying counterpart to WithInsecureSkipVerify: use it to reach
// a hub whose certificate is signed by a private CA.
//
// Because it configures the transport of the client set at the time it runs,
// it must be passed after WithHTTPClient.
func WithRootCAFile(path string) ClientOption {
	return func(c *ClientConfiguration) error {
		pem, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("could not read CA certificate file: %w", err)
		}

		transport, err := c.transport()
		if err != nil {
			return err
		}

		pool := transport.TLSClientConfig.RootCAs
		if pool == nil {
			// start from the system roots so trusting a private CA adds to,
			// rather than replaces, the publicly trusted ones
			if pool, err = x509.SystemCertPool(); err != nil || pool == nil {
				pool = x509.NewCertPool()
			}
		}

		if !pool.AppendCertsFromPEM(pem) {
			return fmt.Errorf("no PEM certificate found in %s", path)
		}

		transport.TLSClientConfig.RootCAs = pool

		return nil
	}
}
