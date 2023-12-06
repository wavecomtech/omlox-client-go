// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"fmt"
)

// LocationProvider defines model for LocationProvider.
//
//easyjson:json
type LocationProvider struct {
	// Must be a valid id for the specific location provider type (e.g. a MAC address of a UWB tag).
	// Wherever applicable, the format of a MAC address SHOULD be an upper case EUI-64 hex-string representation, built from unsigned values. The id MUST including leading zeros and MUST use colon as byte delimiter.
	// IDs which can not be mapped to EUI- 64 MAY deviate from this format.
	ID string `json:"id"`

	// Type of the location provider.
	// A virtual location provider can be used for assigning a unique id to a location provider
	// which does not have a unique identifier by itself.
	// For example, an iOS app will not get the MAC address of the Wi-Fi interface for WiFi positioning.
	// Instead, it will create a virtual location provider to identify the provider and the trackable (iOS device) for location updates.
	Type LocationProviderType `json:"type"`

	// An optional name for the location provider.
	Name string `json:"name,omitempty"`

	// Sensors data related to a provider.
	// The actual structure of the sensors data is application defined.
	Sensors interface{} `json:"sensors,omitempty"`

	// The timeout in milliseconds after which a location should expire
	// and trigger a fence exit event (if no more location updates are sent).
	// Must be a positive number or -1 in case of an infinite timeout.
	// If not set, or set to null, it will default to the trackable or fence setting.
	FenceTimeout Duration `json:"fence_timeout,omitempty"`

	// The minimum distance in meters to release from an ongoing collision or fence event.
	// Must be a positive number. If not set or null exit_tolerance will default to 0.
	ExitTolerance float64 `json:"exit_tolerance,omitempty"`

	// The timeout in milliseconds after which a collision outside of an obstacle but still within exit_tolerance distance
	// should release from a collision or fence event.
	// Must be a positive number or -1 in case of an infinite timeout.
	// If not set, or set to null, it will default to the fence or trackable setting.
	ToleranceTimeout Duration `json:"tolerance_timeout,omitempty"`

	// The delay in milliseconds in which an imminent exit event should wait for another location update.
	// This is relevant for fast rate position updates with quick moving objects.
	// For example, an RTLS provider may batch location updates into groups, resulting in distances being temporarily outdated
	// and premature events between quickly moving objects.
	// The provided number must be positive or -1 in case of an infinite exit_delay.
	// If not set, or set to null, it will default to the fence or trackable setting.
	ExitDelay Duration `json:"exit_delay,omitempty"`

	// Any additional application or vendor specific properties.
	// An application implementing this object is not required to interpret any of the custom properties,
	// but it MUST preserve the properties if set.
	Properties json.RawMessage `json:"properties,omitempty"`
}

// The location provider type which triggered this location update.
type LocationProviderType int

// Defines values for LocationProviderType.
const (
	LocationProviderTypeUnknown LocationProviderType = iota
	LocationProviderTypeUwb
	LocationProviderTypeGps
	LocationProviderTypeWifi
	LocationProviderTypeRfid
	LocationProviderTypeIbeacon
	LocationProviderTypeVirtual
)

// FromString assigs itself from type name.
func (t *LocationProviderType) FromString(name string) error {
	v, ok := map[string]LocationProviderType{
		LocationProviderTypeUnknown.String(): LocationProviderTypeUnknown,
		LocationProviderTypeUwb.String():     LocationProviderTypeUwb,
		LocationProviderTypeGps.String():     LocationProviderTypeGps,
		LocationProviderTypeWifi.String():    LocationProviderTypeWifi,
		LocationProviderTypeRfid.String():    LocationProviderTypeRfid,
		LocationProviderTypeIbeacon.String(): LocationProviderTypeIbeacon,
		LocationProviderTypeVirtual.String(): LocationProviderTypeVirtual,
	}[name]

	if !ok {
		return fmt.Errorf("location provider of type %s not supported", name)
	}

	*t = v
	return nil
}

// String return a text representation.
func (t LocationProviderType) String() string {
	types := [...]string{
		"unknown",
		"uwb",
		"gps",
		"wifi",
		"rfid",
		"ibeacon",
		"virtual",
	}

	if len(types) < int(t) {
		return ""
	}

	return types[t]
}

// MarshalJSON encodes type in to JSON.
func (t LocationProviderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON decodes type from JSON.
func (t *LocationProviderType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	return t.FromString(s)
}
