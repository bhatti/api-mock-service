package oapi

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldStripQuotes(t *testing.T) {
	contents := `{"job":{"add":"{{RandStringArrayMinMax 1 1}}","attributeMap":"{{RandDict}}","completed":"{{RandBool}}","jobId":"{{RandStringMinMax 0 0}}","jobStatus":"{{EnumString PENDING RUNNING SUCCEEDED CANCELED FAILED}}","name":   "{{RandStringMinMax 0 0}}","records":"{{RandNumMinMax 0 0}}","remaining":"{{RandNumMinMax 0 0}}","remove":"{{RandStringArrayMinMax 1 1}}"}}`
	out := stripQuotes([]byte(contents))
	expected := `{"job":{"add":{{RandStringArrayMinMax 1 1}},"attributeMap":{{RandDict}},"completed":{{RandBool}},"jobId":"{{RandStringMinMax 0 0}}","jobStatus":"{{EnumString PENDING RUNNING SUCCEEDED CANCELED FAILED}}","name":   "{{RandStringMinMax 0 0}}","records":{{RandNumMinMax 0 0}},"remaining":{{RandNumMinMax 0 0}},"remove":{{RandStringArrayMinMax 1 1}}}}`
	require.Equal(t, expected, string(out))
}
