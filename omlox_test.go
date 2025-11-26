// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox_test

import (
	"context"
	"log"
	"time"

	"github.com/wavecomtech/omlox-client-go"
)

func ExampleNew() {
	c, err := omlox.New("localhost:8081/v2")
	if err != nil {
		log.Fatal(err)
	}

	trackables, err := c.Trackables.List(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	_ = trackables // use trackables
}

func ExampleConnect() {
	// Dials a Omlox Hub websocket interface, subscribes to
	// the location_updates topic and listens to new
	// location messages.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := omlox.Connect(
		ctx,
		"localhost:7081/v2",
		omlox.WithWSAutoReconnect(true),
		omlox.WithWSMaxRetries(-1), // Unlimited retries
		omlox.WithWSRetryWait(time.Second, 30*time.Second), // Backoff timing
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	sub, err := client.Subscribe(ctx, omlox.TopicLocationUpdates)
	if err != nil {
		log.Fatal(err)
	}

	for location := range omlox.ReceiveAs[omlox.Location](sub) {
		_ = location // handle location update
	}
}
