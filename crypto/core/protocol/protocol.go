//
// Copyright Coinbase, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

var (
	ErrNotInitialized   = fmt.Errorf("object has not been initialized")
	ErrProtocolFinished = fmt.Errorf("the protocol has finished")
)

const (
	// Dkls18Dkg specifies the DKG protocol of the DKLs18 potocol.
	Dkls18Dkg = "DKLs18-DKG"

	// Dkls18Sign specifies the DKG protocol of the DKLs18 potocol.
	Dkls18Sign = "DKLs18-Sign"

	// Dkls18Refresh specifies the DKG protocol of the DKLs18 potocol.
	Dkls18Refresh = "DKLs18-Refresh"

	// versions will increment in 100 intervals, to leave room for adding other versions in between them if it is
	// ever needed in the future.

	// Version0 is version 0!
	Version0 = 100

	// Version1 is version 2!
	Version1 = 200
)

// Message provides serializers and deserializer for the inputs and outputs of each step of the protocol.
// Moreover, it adds some metadata and versioning around the serialized data.
type Message struct {
	Payloads map[string][]byte `json:"payloads"`
	Metadata map[string]string `json:"metadata"`
	Protocol string            `json:"protocol"`
	Version  uint              `json:"version"`
}

// EncodeMessage encodes the message to a string.
func EncodeMessage(m *Message) (string, error) {
	bz, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bz), nil
}

// DecodeMessage decodes the message from a string.
func DecodeMessage(s string) (*Message, error) {
	bz, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var m Message
	if err := json.Unmarshal(bz, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// MarshalJSON marshals the message to JSON.
func (m *Message) MarshalJSON() ([]byte, error) {
	type Alias Message // Use type alias to avoid infinite recursion
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	})
}

// UnmarshalJSON unmarshals the message from JSON.
func (m *Message) UnmarshalJSON(data []byte) error {
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	m.Payloads = make(map[string][]byte)
	m.Metadata = make(map[string]string)
	for k, v := range obj {
		switch k {
		case "payloads":
			m.Payloads = v.(map[string][]byte)
		case "metadata":
			m.Metadata = v.(map[string]string)
		case "protocol":
			m.Protocol = v.(string)
		case "version":
			m.Version = uint(v.(float64))
		}
	}
	return nil
}

// Iterator an interface for the DKLs18 protocols that follows the iterator pattern.
type Iterator interface {
	// Next runs the next round of the protocol.
	// Returns `ErrProtocolFinished` when protocol has completed.
	Next(input *Message) (*Message, error)

	// Result returns the final result, if any, of the completed protocol.
	// Returns nil if the protocol has not yet terminated.
	// Returns an error if an error was encountered during protocol execution.
	Result(version uint) (*Message, error)
}
