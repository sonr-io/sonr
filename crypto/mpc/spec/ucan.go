package spec

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// Token represents a UCAN JWT token with parsed claims
type Token struct {
	Raw          string        `json:"raw"`
	Issuer       string        `json:"iss"`
	Audience     string        `json:"aud"`
	ExpiresAt    int64         `json:"exp,omitempty"`
	NotBefore    int64         `json:"nbf,omitempty"`
	Attenuations []Attenuation `json:"att"`
	Proofs       []Proof       `json:"prf,omitempty"`
	Facts        []Fact        `json:"fct,omitempty"`
}

// Attenuation represents a UCAN capability attenuation
type Attenuation struct {
	Capability Capability `json:"can"`
	Resource   Resource   `json:"with"`
}

// Proof represents a UCAN delegation proof (either JWT or CID)
type Proof string

// Fact represents arbitrary facts in UCAN tokens
type Fact struct {
	Data json.RawMessage `json:"data"`
}

// Capability defines what actions can be performed
type Capability interface {
	GetActions() []string
	Grants(abilities []string) bool
	Contains(other Capability) bool
	String() string
}

// Resource defines what resource the capability applies to
type Resource interface {
	GetScheme() string
	GetValue() string
	GetURI() string
	Matches(other Resource) bool
}

// SimpleCapability implements Capability for single actions
type SimpleCapability struct {
	Action string `json:"action"`
}

func (c *SimpleCapability) GetActions() []string { return []string{c.Action} }
func (c *SimpleCapability) Grants(abilities []string) bool {
	return len(abilities) == 1 && c.Action == abilities[0]
}

func (c *SimpleCapability) Contains(
	other Capability,
) bool {
	return c.Action == other.GetActions()[0]
}
func (c *SimpleCapability) String() string { return c.Action }

// SimpleResource implements Resource for basic URI resources
type SimpleResource struct {
	Scheme string `json:"scheme"`
	Value  string `json:"value"`
	URI    string `json:"uri"`
}

func (r *SimpleResource) GetScheme() string           { return r.Scheme }
func (r *SimpleResource) GetValue() string            { return r.Value }
func (r *SimpleResource) GetURI() string              { return r.URI }
func (r *SimpleResource) Matches(other Resource) bool { return r.URI == other.GetURI() }

// UCAN constants
const (
	UCANVersion    = "0.9.0"
	UCANVersionKey = "ucv"
	PrfKey         = "prf"
	FctKey         = "fct"
	AttKey         = "att"
	CapKey         = "cap"
)

// CreateSimpleAttenuation creates a basic attenuation
func CreateSimpleAttenuation(action, resourceURI string) Attenuation {
	return Attenuation{
		Capability: &SimpleCapability{Action: action},
		Resource:   parseResourceURI(resourceURI),
	}
}

// parseResourceURI creates a Resource from URI string
func parseResourceURI(uri string) Resource {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return &SimpleResource{
			Scheme: "unknown",
			Value:  uri,
			URI:    uri,
		}
	}

	return &SimpleResource{
		Scheme: parts[0],
		Value:  parts[1],
		URI:    uri,
	}
}

// getIssuerDIDFromBytes creates an issuer DID and address from public key bytes (alternative implementation)
func getIssuerDIDFromBytesAlt(pubKeyBytes []byte) (string, string, error) {
	addr, err := bech32.ConvertAndEncode("idx", pubKeyBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode address: %w", err)
	}
	return fmt.Sprintf("did:sonr:%s", addr), addr, nil
}
