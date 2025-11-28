package sql

import (
	"terraform-provider-azuresql/internal/logging"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPairwiseReverseHex(t *testing.T) {
	assert.Equal(t, PairwiseReverseHex(15, 1), "0F", "PairwiseReverseHex(15, 1) should be 0F.")
	assert.Equal(t, PairwiseReverseHex(2, 1), "02", "PairwiseReverseHex(2, 1) should be 02.")
	assert.Equal(t, PairwiseReverseHex(3545725711, 4), "0F7B57D3", "PairwiseReverseHex(3545725711, 4) should be 0F.")
}

func TestAzureSIDToDatabaseSID(t *testing.T) {
	ctx := logging.GetTestContext()
	assert.Equal(t, AzureSIDToDatabaseSID(ctx, "S-1-12-1-165585138-1090985625-2859435433-1111111111"), "0xF2A0DE09991E0741A9856FAAC7353A42")
}

func TestObjectIDToDatabaseSID(t *testing.T) {
	ctx := logging.GetTestContext()
	assert.Equal(t, ObjectIDToDatabaseSID(ctx, "09dea0f2-1e99-4107-a985-111111111111"), "0xF2A0DE09991E0741A985111111111111")
}
