// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

// tlsHubServer is an Omlox Hub websocket interface served over TLS with a
// self-signed certificate, as a hub behind a reverse proxy with no CA-issued
// certificate presents.
func tlsHubServer(t *testing.T) *httptest.Server {
	t.Helper()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		<-r.Context().Done()
	}))

	t.Cleanup(srv.Close)

	return srv
}

// caFile writes the server's own certificate out as a PEM file, so it can be
// trusted as a root the way a private CA certificate would be.
func caFile(t *testing.T, srv *httptest.Server) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "ca.pem")
	encoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})

	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("writing CA file: %v", err)
	}

	return path
}

func TestConnectOverTLS(t *testing.T) {
	srv := tlsHubServer(t)

	tests := []struct {
		name    string
		options func(t *testing.T) []ClientOption
		wantErr string
	}{
		{
			name:    "untrusted certificate is rejected by default",
			options: func(*testing.T) []ClientOption { return nil },
			wantErr: "certificate signed by unknown authority",
		},
		{
			name: "skipping verification accepts it",
			options: func(*testing.T) []ClientOption {
				return []ClientOption{WithInsecureSkipVerify(true)}
			},
		},
		{
			name: "trusting the CA accepts it while still verifying",
			options: func(t *testing.T) []ClientOption {
				return []ClientOption{WithRootCAFile(caFile(t, srv))}
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			client, err := Connect(ctx, srv.URL+"/v2", tc.options(t)...)
			if client != nil {
				defer client.Close()
			}

			switch {
			case tc.wantErr == "":
				if err != nil {
					t.Fatalf("connecting: %v", err)
				}
			case err == nil:
				t.Fatalf("connected, want error containing %q", tc.wantErr)
			case !strings.Contains(err.Error(), tc.wantErr):
				t.Fatalf("error = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

func TestWithRootCAFile(t *testing.T) {
	t.Run("reports a missing file", func(t *testing.T) {
		_, err := New("https://hub.invalid/v2", WithRootCAFile(filepath.Join(t.TempDir(), "absent.pem")))
		if err == nil || !strings.Contains(err.Error(), "could not read CA certificate file") {
			t.Fatalf("error = %v, want it to report the unreadable file", err)
		}
	})

	t.Run("reports a file holding no certificate", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "empty.pem")
		if err := os.WriteFile(path, []byte("not a certificate\n"), 0o600); err != nil {
			t.Fatalf("writing file: %v", err)
		}

		_, err := New("https://hub.invalid/v2", WithRootCAFile(path))
		if err == nil || !strings.Contains(err.Error(), "no PEM certificate found") {
			t.Fatalf("error = %v, want it to report the missing certificate", err)
		}
	})
}

// wrappedTransport stands in for a retrying or instrumenting RoundTripper of
// the kind callers wrap the HTTP client with.
type wrappedTransport struct{ http.RoundTripper }

// TestTLSOptionsRejectAWrappedTransport documents why a caller that wraps the
// HTTP client — as the RTLS adapters do to retry API calls — has to configure
// TLS on the inner client instead: the wrapper hides the *http.Transport the
// options need to reach.
func TestTLSOptionsRejectAWrappedTransport(t *testing.T) {
	wrapped := &http.Client{Transport: &wrappedTransport{http.DefaultTransport}}

	_, err := New("https://hub.invalid/v2", WithHTTPClient(wrapped), WithInsecureSkipVerify(true))
	if err == nil {
		t.Fatal("configuring TLS on a wrapped transport succeeded, want an error")
	}
	if !strings.Contains(err.Error(), "must be *http.Transport") {
		t.Fatalf("error = %v, want it to name the transport requirement", err)
	}
}
