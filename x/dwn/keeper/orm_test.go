package keeper_test

import (
	"testing"

	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	"github.com/stretchr/testify/require"
)

func TestORM(t *testing.T) {
	f := SetupTest(t)

	// Simple ORM test
	recordTable := f.k.OrmDB.DWNRecordTable()
	record := &apiv1.DWNRecord{
		RecordId: "test-record-123",
		Target:   "did:example:123",
	}

	err := recordTable.Insert(f.ctx, record)
	require.NoError(t, err)

	has, err := recordTable.Has(f.ctx, record.RecordId)
	require.NoError(t, err)
	require.True(t, has)
}
