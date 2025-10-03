package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sonr-io/sonr/x/did/types"
)

type EventsTestSuite struct {
	suite.Suite
	f *testFixture
}

func TestEventsTestSuite(t *testing.T) {
	suite.Run(t, new(EventsTestSuite))
}

func (suite *EventsTestSuite) SetupTest() {
	suite.f = SetupTest(suite.T())
}

// TestCreateDIDEventEmission tests that EventDIDCreated is properly emitted
func (suite *EventsTestSuite) TestCreateDIDEventEmission() {
	did := "did:sonr:testuser123"
	controller := suite.f.addrs[0].String()

	msg := &types.MsgCreateDID{
		Controller: controller,
		DidDocument: types.DIDDocument{
			Id:                did,
			PrimaryController: controller,
			VerificationMethod: []*types.VerificationMethod{
				{
					Id:                     did + "#key-1",
					VerificationMethodKind: "Ed25519VerificationKey2020",
					Controller:             did,
					PublicKeyMultibase:     "zH3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
				},
			},
			Service: []*types.Service{
				{
					Id:             did + "#service-1",
					ServiceKind:    "LinkedDomains",
					SingleEndpoint: "https://example.com",
				},
			},
		},
	}

	// Execute CreateDID
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, msg)
	suite.Require().NoError(err)

	// Check for emitted events
	events := suite.f.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Find the typed event
	var foundEvent bool
	for _, event := range events {
		if event.Type == "did.v1.EventDIDCreated" {
			foundEvent = true
			break
		}
	}

	suite.Require().True(foundEvent, "EventDIDCreated not found in emitted events")
}

// TestUpdateDIDEventEmission tests that EventDIDUpdated is properly emitted
func (suite *EventsTestSuite) TestUpdateDIDEventEmission() {
	did := "did:sonr:testuser456"
	controller := suite.f.addrs[0].String()

	// First create the DID
	createMsg := &types.MsgCreateDID{
		Controller: controller,
		DidDocument: types.DIDDocument{
			Id:                did,
			PrimaryController: controller,
		},
	}

	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, createMsg)
	suite.Require().NoError(err)

	// Clear events from creation
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Now update the DID
	updateMsg := &types.MsgUpdateDID{
		Did:        did,
		Controller: controller,
		DidDocument: types.DIDDocument{
			Id:                did,
			PrimaryController: controller,
			Service: []*types.Service{
				{
					Id:             did + "#new-service",
					ServiceKind:    "LinkedDomains",
					SingleEndpoint: "https://updated.com",
				},
			},
		},
	}

	_, err = suite.f.msgServer.UpdateDID(suite.f.ctx, updateMsg)
	suite.Require().NoError(err)

	// Check for emitted events
	events := suite.f.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Verify EventDIDUpdated was emitted
	var foundEvent bool
	for _, event := range events {
		if event.Type == "did.v1.EventDIDUpdated" {
			foundEvent = true
			break
		}
	}

	suite.Require().True(foundEvent, "EventDIDUpdated not found in emitted events")
}

// TestDeactivateDIDEventEmission tests that EventDIDDeactivated is properly emitted
func (suite *EventsTestSuite) TestDeactivateDIDEventEmission() {
	did := "did:sonr:testuser789"
	controller := suite.f.addrs[0].String()

	// First create the DID
	createMsg := &types.MsgCreateDID{
		Controller: controller,
		DidDocument: types.DIDDocument{
			Id:                did,
			PrimaryController: controller,
		},
	}

	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, createMsg)
	suite.Require().NoError(err)

	// Clear events from creation
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Now deactivate the DID
	deactivateMsg := &types.MsgDeactivateDID{
		Did:        did,
		Controller: controller,
	}

	_, err = suite.f.msgServer.DeactivateDID(suite.f.ctx, deactivateMsg)
	suite.Require().NoError(err)

	// Check for emitted events
	events := suite.f.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Verify EventDIDDeactivated was emitted
	var foundEvent bool
	for _, event := range events {
		if event.Type == "did.v1.EventDIDDeactivated" {
			foundEvent = true
			break
		}
	}

	suite.Require().True(foundEvent, "EventDIDDeactivated not found in emitted events")
}

// TestErrorCaseNoEventEmission tests that events are not emitted on errors
func (suite *EventsTestSuite) TestErrorCaseNoEventEmission() {
	// Try to create an invalid DID
	msg := &types.MsgCreateDID{
		Controller: suite.f.addrs[0].String(),
		DidDocument: types.DIDDocument{
			Id: "", // Invalid empty ID
		},
	}

	// Clear any previous events
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Execute CreateDID - should fail
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, msg)
	suite.Require().Error(err)

	// Check that no events were emitted (except potentially message events)
	events := suite.f.ctx.EventManager().Events()

	// Filter out message events
	var nonMessageEvents []sdk.Event
	for _, event := range events {
		if event.Type != sdk.EventTypeMessage {
			nonMessageEvents = append(nonMessageEvents, event)
		}
	}

	suite.Require().Empty(nonMessageEvents, "Expected no events to be emitted on error")
}
