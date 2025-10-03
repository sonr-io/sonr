package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	svcv1 "github.com/sonr-io/sonr/api/svc/v1"
	"github.com/sonr-io/sonr/x/svc/types"
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

// TestInitiateDomainVerificationEventEmission tests EventDomainVerificationInitiated
func (suite *EventsTestSuite) TestInitiateDomainVerificationEventEmission() {
	domain := "example.com"
	creator := suite.f.addrs[0].String()

	msg := &types.MsgInitiateDomainVerification{
		Domain:  domain,
		Creator: creator,
	}

	// Execute InitiateDomainVerification
	resp, err := suite.f.msgServer.InitiateDomainVerification(suite.f.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().NotEmpty(resp.VerificationToken)

	// Check for emitted events
	events := suite.f.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Find the typed event - simplified check
	var foundEvent bool
	for _, event := range events {
		if event.Type == "svc.v1.EventDomainVerificationInitiated" {
			foundEvent = true
			break
		}
	}

	suite.Require().True(foundEvent, "EventDomainVerificationInitiated not found")
}

// TestVerifyDomainEventEmission tests EventDomainVerified emission
func (suite *EventsTestSuite) TestVerifyDomainEventEmission() {
	domain := "verified.com"
	creator := suite.f.addrs[0].String()

	// First initiate verification
	initMsg := &types.MsgInitiateDomainVerification{
		Domain:  domain,
		Creator: creator,
	}

	initResp, err := suite.f.msgServer.InitiateDomainVerification(suite.f.ctx, initMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(initResp)

	// Mock successful DNS verification by updating the verification status directly
	// In real scenario, DNS would be checked
	verification, err := suite.f.k.OrmDB.DomainVerificationTable().Get(suite.f.ctx, domain)
	suite.Require().NoError(err)
	suite.Require().NotNil(verification)

	// Update status to verified (simulating successful DNS check)
	verification.Status = svcv1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED
	err = suite.f.k.OrmDB.DomainVerificationTable().Update(suite.f.ctx, verification)
	suite.Require().NoError(err)

	// Clear events
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Now verify the domain
	verifyMsg := &types.MsgVerifyDomain{
		Domain:  domain,
		Creator: creator,
	}

	verifyResp, err := suite.f.msgServer.VerifyDomain(suite.f.ctx, verifyMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(verifyResp)
	suite.Require().True(verifyResp.Verified)

	// Check for emitted events
	events := suite.f.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Find the EventDomainVerified - simplified check
	var foundEvent bool
	for _, event := range events {
		if event.Type == "svc.v1.EventDomainVerified" {
			foundEvent = true
			break
		}
	}

	suite.Require().True(foundEvent, "EventDomainVerified not found")
}

// TestRegisterServiceEventEmission tests EventServiceRegistered emission
func (suite *EventsTestSuite) TestRegisterServiceEventEmission() {
	domain := "service.com"
	creator := suite.f.addrs[0].String()
	serviceId := "test-service-001"

	// Setup: Create and verify domain first
	suite.setupVerifiedDomain(domain, creator)

	// Clear events
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Register service
	msg := &types.MsgRegisterService{
		ServiceId:            serviceId,
		Domain:               domain,
		Creator:              creator,
		RequestedPermissions: []string{"read", "write"},
	}

	resp, err := suite.f.msgServer.RegisterService(suite.f.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Equal(serviceId, resp.ServiceId)

	// Check for emitted events
	events := suite.f.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Find the EventServiceRegistered - simplified check
	var foundEvent bool
	for _, event := range events {
		if event.Type == "svc.v1.EventServiceRegistered" {
			foundEvent = true
			break
		}
	}

	suite.Require().True(foundEvent, "EventServiceRegistered not found")
}

// TestFailedVerificationNoEventEmission tests no event on failed verification
func (suite *EventsTestSuite) TestFailedVerificationNoEventEmission() {
	domain := "unverified.com"
	creator := suite.f.addrs[0].String()

	// Initiate verification but don't set DNS record
	initMsg := &types.MsgInitiateDomainVerification{
		Domain:  domain,
		Creator: creator,
	}

	_, err := suite.f.msgServer.InitiateDomainVerification(suite.f.ctx, initMsg)
	suite.Require().NoError(err)

	// Clear events
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Try to verify without DNS record (should fail)
	verifyMsg := &types.MsgVerifyDomain{
		Domain:  domain,
		Creator: creator,
	}

	resp, err := suite.f.msgServer.VerifyDomain(suite.f.ctx, verifyMsg)
	suite.Require().NoError(err) // Returns success with verified=false
	suite.Require().False(resp.Verified)

	// Check that no EventDomainVerified was emitted
	events := suite.f.ctx.EventManager().Events()

	// Look for EventDomainVerified - should not find it
	var foundVerifiedEvent bool
	for _, event := range events {
		if event.Type == "svc.v1.EventDomainVerified" {
			foundVerifiedEvent = true
			break
		}
	}

	suite.Require().
		False(foundVerifiedEvent, "EventDomainVerified should not be emitted on failure")
}

// TestErrorCaseNoEventEmission tests no events on error
func (suite *EventsTestSuite) TestErrorCaseNoEventEmission() {
	// Try to register service without verified domain
	msg := &types.MsgRegisterService{
		ServiceId:            "invalid-service",
		Domain:               "notverified.com",
		Creator:              suite.f.addrs[0].String(),
		RequestedPermissions: []string{"read"},
	}

	// Clear any previous events
	suite.f.ctx = suite.f.ctx.WithEventManager(sdk.NewEventManager())

	// Execute RegisterService - should fail
	_, err := suite.f.msgServer.RegisterService(suite.f.ctx, msg)
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

// Helper function to setup a verified domain
func (suite *EventsTestSuite) setupVerifiedDomain(domain, creator string) {
	// Initiate verification
	initMsg := &types.MsgInitiateDomainVerification{
		Domain:  domain,
		Creator: creator,
	}

	_, err := suite.f.msgServer.InitiateDomainVerification(suite.f.ctx, initMsg)
	suite.Require().NoError(err)

	// Mock successful verification
	verification, err := suite.f.k.OrmDB.DomainVerificationTable().Get(suite.f.ctx, domain)
	suite.Require().NoError(err)
	suite.Require().NotNil(verification)

	verification.Status = svcv1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED
	err = suite.f.k.OrmDB.DomainVerificationTable().Update(suite.f.ctx, verification)
	suite.Require().NoError(err)
}
