package keeper_test

import (
	"testing"
	"time"

	apiv1 "github.com/sonr-io/sonr/api/svc/v1"
	"github.com/stretchr/testify/require"
)

func TestDomainVerificationORM(t *testing.T) {
	f := SetupTest(t)

	dt := f.k.OrmDB.DomainVerificationTable()
	domain := "example.com"
	owner := "cosmos1abc123"
	token := "verification-token-12345"
	now := time.Now().Unix()

	// Test Insert
	err := dt.Insert(f.ctx, &apiv1.DomainVerification{
		Domain:            domain,
		Owner:             owner,
		VerificationToken: token,
		Status:            apiv1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_PENDING,
		ExpiresAt:         now + 3600, // 1 hour from now
		VerifiedAt:        0,          // Not verified yet
	})
	require.NoError(t, err)

	// Test Has
	exists, err := dt.Has(f.ctx, domain)
	require.NoError(t, err)
	require.True(t, exists)

	// Test Get
	res, err := dt.Get(f.ctx, domain)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, domain, res.Domain)
	require.Equal(t, owner, res.Owner)
	require.Equal(t, token, res.VerificationToken)
	require.Equal(t, apiv1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_PENDING, res.Status)
	require.Equal(t, now+3600, res.ExpiresAt)
	require.Equal(t, int64(0), res.VerifiedAt)

	// Test Update
	res.Status = apiv1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED
	res.VerifiedAt = now
	err = dt.Update(f.ctx, res)
	require.NoError(t, err)

	// Verify update
	updated, err := dt.Get(f.ctx, domain)
	require.NoError(t, err)
	require.Equal(
		t,
		apiv1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED,
		updated.Status,
	)
	require.Equal(t, now, updated.VerifiedAt)
}

func TestServiceORM(t *testing.T) {
	f := SetupTest(t)

	st := f.k.OrmDB.ServiceTable()
	serviceID := "service-123"
	domain := "api.example.com"
	owner := "cosmos1def456"
	capabilityCID := "QmServiceCapability123"
	now := time.Now().Unix()

	// Test Insert
	err := st.Insert(f.ctx, &apiv1.Service{
		Id:                serviceID,
		Domain:            domain,
		Owner:             owner,
		RootCapabilityCid: capabilityCID,
		Permissions:       []string{"register", "update"},
		Status:            apiv1.ServiceStatus_SERVICE_STATUS_ACTIVE,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	require.NoError(t, err)

	// Test Has
	exists, err := st.Has(f.ctx, serviceID)
	require.NoError(t, err)
	require.True(t, exists)

	// Test Get
	res, err := st.Get(f.ctx, serviceID)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, serviceID, res.Id)
	require.Equal(t, domain, res.Domain)
	require.Equal(t, owner, res.Owner)
	require.Equal(t, capabilityCID, res.RootCapabilityCid)
	require.Equal(t, []string{"register", "update"}, res.Permissions)
	require.Equal(t, apiv1.ServiceStatus_SERVICE_STATUS_ACTIVE, res.Status)
	require.Equal(t, now, res.CreatedAt)
	require.Equal(t, now, res.UpdatedAt)

	// Test Update
	res.Status = apiv1.ServiceStatus_SERVICE_STATUS_SUSPENDED
	res.UpdatedAt = now + 100
	err = st.Update(f.ctx, res)
	require.NoError(t, err)

	// Verify update
	updated, err := st.Get(f.ctx, serviceID)
	require.NoError(t, err)
	require.Equal(t, apiv1.ServiceStatus_SERVICE_STATUS_SUSPENDED, updated.Status)
	require.Equal(t, now+100, updated.UpdatedAt)
}

func TestServiceIndexQueries(t *testing.T) {
	f := SetupTest(t)

	st := f.k.OrmDB.ServiceTable()
	owner := "cosmos1test123"
	domain := "test.example.com"
	now := time.Now().Unix()

	// Insert test services
	services := []*apiv1.Service{
		{
			Id:                "service-1",
			Domain:            domain,
			Owner:             owner,
			RootCapabilityCid: "QmCap1",
			Permissions:       []string{"register"},
			Status:            apiv1.ServiceStatus_SERVICE_STATUS_ACTIVE,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Id:                "service-2",
			Domain:            "other.example.com",
			Owner:             owner,
			RootCapabilityCid: "QmCap2",
			Permissions:       []string{"update"},
			Status:            apiv1.ServiceStatus_SERVICE_STATUS_SUSPENDED,
			CreatedAt:         now + 100,
			UpdatedAt:         now + 100,
		},
	}

	for _, service := range services {
		err := st.Insert(f.ctx, service)
		require.NoError(t, err)
	}

	// Test query by owner index
	ownerKey := apiv1.ServiceOwnerIndexKey{}.WithOwner(owner)
	iter, err := st.List(f.ctx, ownerKey)
	require.NoError(t, err)

	var ownerServices []*apiv1.Service
	for iter.Next() {
		service, errb := iter.Value()
		require.NoError(t, errb)
		ownerServices = append(ownerServices, service)
	}
	iter.Close()

	require.Len(t, ownerServices, 2)

	// Test query by domain index
	domainKey := apiv1.ServiceDomainIndexKey{}.WithDomain(domain)
	iter, err = st.List(f.ctx, domainKey)
	require.NoError(t, err)

	var domainServices []*apiv1.Service
	for iter.Next() {
		service, err := iter.Value()
		require.NoError(t, err)
		domainServices = append(domainServices, service)
	}
	iter.Close()

	require.Len(t, domainServices, 1)
	require.Equal(t, "service-1", domainServices[0].Id)
}
