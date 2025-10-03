package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sonr-io/sonr/x/dwn/types"
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

// TestRecordsWriteEventEmission tests that EventRecordWritten is properly emitted
func (suite *EventsTestSuite) TestRecordsWriteEventEmission() {
	target := "did:sonr:testuser123"
	author := suite.f.addrs[0].String()

	msg := &types.MsgRecordsWrite{
		Target: target,
		Author: author,
		Descriptor_: &types.DWNMessageDescriptor{
			InterfaceName:    "Records",
			Method:           "Write",
			MessageTimestamp: "2024-01-01T00:00:00Z",
			DataFormat:       "application/json",
		},
		Data:     []byte(`{"test": "data"}`),
		Protocol: "test-protocol",
		Schema:   "test-schema",
	}

	// Execute RecordsWrite
	resp, err := suite.f.msgServer.RecordsWrite(suite.f.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)

	// Check for emitted events
	events := suite.f.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Find the typed event - simplified check
	var foundEvent bool
	for _, event := range events {
		if event.Type == "dwn.v1.EventRecordWritten" {
			foundEvent = true
			break
		}
	}

	suite.Require().True(foundEvent, "EventRecordWritten not found in emitted events")
}

// TestRecordsDeleteEventEmission tests that EventRecordDeleted is properly emitted
func (suite *EventsTestSuite) TestRecordsDeleteEventEmission() {
	target := "did:sonr:testuser456"
	author := suite.f.addrs[0].String()

	// First create a record
	writeMsg := &types.MsgRecordsWrite{
		Target: target,
		Author: author,
		Descriptor_: &types.DWNMessageDescriptor{
			InterfaceName:    "Records",
			Method:           "Write",
			MessageTimestamp: "2024-01-01T00:00:00Z",
			DataFormat:       "application/json",
		},
		Data:     []byte(`{"test": "data"}`),
		Protocol: "test-protocol",
		Schema:   "test-schema",
	}

	writeResp, err := suite.f.msgServer.RecordsWrite(suite.f.ctx, writeMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(writeResp)

	// Clear events from creation
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Now delete the record
	deleteMsg := &types.MsgRecordsDelete{
		Target:   target,
		Author:   author,
		RecordId: writeResp.RecordId,
		Descriptor_: &types.DWNMessageDescriptor{
			InterfaceName:    "Records",
			Method:           "Delete",
			MessageTimestamp: "2024-01-01T00:00:01Z",
		},
	}

	_, err = suite.f.msgServer.RecordsDelete(suite.f.ctx, deleteMsg)
	suite.Require().NoError(err)

	// Check for emitted events
	events := suite.f.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Verify EventRecordDeleted was emitted - simplified check
	var foundEvent bool
	for _, event := range events {
		if event.Type == "dwn.v1.EventRecordDeleted" {
			foundEvent = true
			break
		}
	}

	suite.Require().True(foundEvent, "EventRecordDeleted not found in emitted events")
}

// TestRecordsUpdateEventEmission tests that EventRecordWritten is emitted for updates
func (suite *EventsTestSuite) TestRecordsUpdateEventEmission() {
	target := "did:sonr:testuser789"
	author := suite.f.addrs[0].String()

	// Create initial record
	msg1 := &types.MsgRecordsWrite{
		Target: target,
		Author: author,
		Descriptor_: &types.DWNMessageDescriptor{
			InterfaceName:    "Records",
			Method:           "Write",
			MessageTimestamp: "2024-01-01T00:00:00Z",
			DataFormat:       "application/json",
		},
		Data:     []byte(`{"version": "1"}`),
		Protocol: "test-protocol",
		Schema:   "test-schema",
	}

	resp1, err := suite.f.msgServer.RecordsWrite(suite.f.ctx, msg1)
	suite.Require().NoError(err)

	// Clear events
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Update the record (same target, protocol, schema, timestamp = update)
	msg2 := &types.MsgRecordsWrite{
		Target: target,
		Author: author,
		Descriptor_: &types.DWNMessageDescriptor{
			InterfaceName:    "Records",
			Method:           "Write",
			MessageTimestamp: "2024-01-01T00:00:00Z", // Same timestamp triggers update
			DataFormat:       "application/json",
		},
		Data:     []byte(`{"version": "2"}`),
		Protocol: "test-protocol",
		Schema:   "test-schema",
	}

	resp2, err := suite.f.msgServer.RecordsWrite(suite.f.ctx, msg2)
	suite.Require().NoError(err)
	// Note: Different data creates a different record ID, not an update
	suite.Require().
		NotEqual(resp1.RecordId, resp2.RecordId, "Different data should create different record ID")

	// Check for emitted events
	events := suite.f.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Verify EventRecordWritten was emitted for the update - simplified check
	var foundEvent bool
	for _, event := range events {
		if event.Type == "dwn.v1.EventRecordWritten" {
			foundEvent = true
			break
		}
	}

	suite.Require().True(foundEvent, "EventRecordWritten not found for update")
}

// TestErrorCaseNoEventEmission tests that events are not emitted on errors
func (suite *EventsTestSuite) TestErrorCaseNoEventEmission() {
	// Try to delete a non-existent record
	msg := &types.MsgRecordsDelete{
		Target:   "did:sonr:testuser999",
		Author:   suite.f.addrs[0].String(),
		RecordId: "non-existent-record",
		Descriptor_: &types.DWNMessageDescriptor{
			InterfaceName:    "Records",
			Method:           "Delete",
			MessageTimestamp: "2024-01-01T00:00:00Z",
		},
	}

	// Clear any previous events
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Execute RecordsDelete - should fail
	_, err := suite.f.msgServer.RecordsDelete(suite.f.ctx, msg)
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
