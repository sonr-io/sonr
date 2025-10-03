package client

import (
	"context"
	"fmt"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

// ChannelResponse represents IBC channel query response
type ChannelResponse struct {
	Channel     channeltypes.Channel `json:"channel"`
	Proof       []byte               `json:"proof"`
	ProofHeight struct {
		RevisionNumber string `json:"revision_number"`
		RevisionHeight string `json:"revision_height"`
	} `json:"proof_height"`
}

// ChannelsResponse represents IBC channels query response
type ChannelsResponse struct {
	Channels []struct {
		State        string `json:"state"`
		Ordering     string `json:"ordering"`
		Counterparty struct {
			PortID    string `json:"port_id"`
			ChannelID string `json:"channel_id"`
		} `json:"counterparty"`
		ConnectionHops []string `json:"connection_hops"`
		Version        string   `json:"version"`
		PortID         string   `json:"port_id"`
		ChannelID      string   `json:"channel_id"`
	} `json:"channels"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
	Height struct {
		RevisionNumber string `json:"revision_number"`
		RevisionHeight string `json:"revision_height"`
	} `json:"height"`
}

// GetChannel queries an IBC channel
func (c *StarshipClient) GetChannel(ctx context.Context, portID, channelID string) (*ChannelResponse, error) {
	url := fmt.Sprintf("%s/ibc/core/channel/v1/channels/%s/ports/%s", c.baseURL, channelID, portID)

	var channelResp ChannelResponse
	if err := c.doRequest(ctx, url, &channelResp); err != nil {
		return nil, fmt.Errorf("failed to query channel: %w", err)
	}

	return &channelResp, nil
}

// GetChannels queries all IBC channels
func (c *StarshipClient) GetChannels(ctx context.Context) (*ChannelsResponse, error) {
	url := fmt.Sprintf("%s/ibc/core/channel/v1/channels", c.baseURL)

	var channelsResp ChannelsResponse
	if err := c.doRequest(ctx, url, &channelsResp); err != nil {
		return nil, fmt.Errorf("failed to query channels: %w", err)
	}

	return &channelsResp, nil
}

// GetTransferChannel finds the first open transfer channel
func (c *StarshipClient) GetTransferChannel(ctx context.Context) (string, error) {
	channels, err := c.GetChannels(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get channels: %w", err)
	}

	for _, channel := range channels.Channels {
		if channel.PortID == "transfer" && channel.State == "STATE_OPEN" {
			return channel.ChannelID, nil
		}
	}

	return "", fmt.Errorf("no open transfer channel found")
}

// ConnectionResponse represents IBC connection query response
type ConnectionResponse struct {
	Connection struct {
		ClientID string `json:"client_id"`
		Versions []struct {
			Identifier string   `json:"identifier"`
			Features   []string `json:"features"`
		} `json:"versions"`
		State        string `json:"state"`
		Counterparty struct {
			ClientID     string `json:"client_id"`
			ConnectionID string `json:"connection_id"`
			Prefix       struct {
				KeyPrefix []byte `json:"key_prefix"`
			} `json:"prefix"`
		} `json:"counterparty"`
		DelayPeriod string `json:"delay_period"`
	} `json:"connection"`
	Proof       []byte `json:"proof"`
	ProofHeight struct {
		RevisionNumber string `json:"revision_number"`
		RevisionHeight string `json:"revision_height"`
	} `json:"proof_height"`
}

// GetConnection queries an IBC connection
func (c *StarshipClient) GetConnection(ctx context.Context, connectionID string) (*ConnectionResponse, error) {
	url := fmt.Sprintf("%s/ibc/core/connection/v1/connections/%s", c.baseURL, connectionID)

	var connResp ConnectionResponse
	if err := c.doRequest(ctx, url, &connResp); err != nil {
		return nil, fmt.Errorf("failed to query connection: %w", err)
	}

	return &connResp, nil
}

// ClientStateResponse represents IBC client state query response
type ClientStateResponse struct {
	ClientState ibcexported.ClientState `json:"client_state"`
	Proof       []byte                  `json:"proof"`
	ProofHeight struct {
		RevisionNumber string `json:"revision_number"`
		RevisionHeight string `json:"revision_height"`
	} `json:"proof_height"`
}

// GetClientState queries an IBC client state
func (c *StarshipClient) GetClientState(ctx context.Context, clientID string) (*ClientStateResponse, error) {
	url := fmt.Sprintf("%s/ibc/core/client/v1/client_states/%s", c.baseURL, clientID)

	var clientResp ClientStateResponse
	if err := c.doRequest(ctx, url, &clientResp); err != nil {
		return nil, fmt.Errorf("failed to query client state: %w", err)
	}

	return &clientResp, nil
}

// DenomTraceResponse represents IBC denom trace query response
type DenomTraceResponse struct {
	DenomTrace struct {
		Path      string `json:"path"`
		BaseDenom string `json:"base_denom"`
	} `json:"denom_trace"`
}

// GetDenomTrace queries an IBC denom trace
func (c *StarshipClient) GetDenomTrace(ctx context.Context, hash string) (*DenomTraceResponse, error) {
	url := fmt.Sprintf("%s/ibc/apps/transfer/v1/denom_traces/%s", c.baseURL, hash)

	var traceResp DenomTraceResponse
	if err := c.doRequest(ctx, url, &traceResp); err != nil {
		return nil, fmt.Errorf("failed to query denom trace: %w", err)
	}

	return &traceResp, nil
}
