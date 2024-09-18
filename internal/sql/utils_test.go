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
	assert.Equal(t, AzureSIDToDatabaseSID(ctx, "S-1-12-1-165585138-1090985625-2859435433-3545725711"), "0xF2A0DE09991E0741A9856FAA0F7B57D3", "TestAzureSIDToDatabaseSID('S-1-12-1-165585138-1090985625-2859435433-3545725711') should be 0xF2A0DE09991E0741A9856FAA0F7B57D3.")
}
