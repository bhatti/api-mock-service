package contract

import (
	"fmt"
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
)

func Test_ErrorDetails_PopulatedOnStatusMismatch(t *testing.T) {
	resp := types.NewProducerContractResponse()
	cve := &ContractValidationError{
		OriginalError: fmt.Errorf("status 404 != 200"),
		DiffReport: &ContractDiffReport{
			MissingFields:    []string{},
			ExtraFields:      []string{},
			TypeMismatches:   map[string]string{},
			ValueMismatches:  map[string]ValueMismatch{},
			HeaderMismatches: map[string]ValueMismatch{},
			ExpectedFields:   map[string]interface{}{},
			ActualFields:     map[string]interface{}{},
		},
		Scenario: "my-scenario",
		URL:      "https://example.com/api",
	}

	resp.SetErrorDetail("my-scenario_0", buildContractValidationDetail(cve))

	detail := resp.ErrorDetails["my-scenario_0"]
	require.NotNil(t, detail)
	require.Equal(t, "my-scenario", detail.Scenario)
	require.Equal(t, "https://example.com/api", detail.URL)
	require.Contains(t, detail.Summary, "my-scenario")
}

func Test_ErrorDetails_MissingFieldsListed(t *testing.T) {
	resp := types.NewProducerContractResponse()
	cve := &ContractValidationError{
		OriginalError: fmt.Errorf("assertion failed"),
		DiffReport: &ContractDiffReport{
			MissingFields:    []string{"orderId", "status"},
			ExtraFields:      []string{},
			TypeMismatches:   map[string]string{},
			ValueMismatches:  map[string]ValueMismatch{},
			HeaderMismatches: map[string]ValueMismatch{},
			ExpectedFields:   map[string]interface{}{},
			ActualFields:     map[string]interface{}{},
		},
		Scenario: "order-scenario",
		URL:      "https://example.com/orders",
	}

	resp.SetErrorDetail("order-scenario_0", buildContractValidationDetail(cve))

	detail := resp.ErrorDetails["order-scenario_0"]
	require.NotNil(t, detail)
	require.Contains(t, detail.MissingFields, "orderId")
	require.Contains(t, detail.MissingFields, "status")
}

func Test_ErrorDetails_ValueMismatchListed(t *testing.T) {
	resp := types.NewProducerContractResponse()
	cve := &ContractValidationError{
		OriginalError: fmt.Errorf("value mismatch"),
		DiffReport: &ContractDiffReport{
			MissingFields:  []string{},
			ExtraFields:    []string{},
			TypeMismatches: map[string]string{},
			ValueMismatches: map[string]ValueMismatch{
				"price": {Expected: "__number__", Actual: "free"},
			},
			HeaderMismatches: map[string]ValueMismatch{},
			ExpectedFields:   map[string]interface{}{},
			ActualFields:     map[string]interface{}{},
		},
		Scenario: "price-scenario",
		URL:      "https://example.com/prices",
	}

	resp.SetErrorDetail("price-scenario_0", buildContractValidationDetail(cve))

	detail := resp.ErrorDetails["price-scenario_0"]
	require.NotNil(t, detail)
	require.Contains(t, detail.ValueMismatches, "price")
	require.Equal(t, "__number__", detail.ValueMismatches["price"].Expected)
	require.Equal(t, "free", detail.ValueMismatches["price"].Actual)
}

func Test_ErrorDetails_BackwardCompatible(t *testing.T) {
	// The flat Errors map should still be populated as before
	resp := types.NewProducerContractResponse()
	errMsg := fmt.Errorf("something went wrong")
	resp.Add("scenario_0", nil, errMsg)

	require.Contains(t, resp.Errors, "scenario_0")
	require.Equal(t, errMsg.Error(), resp.Errors["scenario_0"])
	// ErrorDetails is NOT set when we use Add() without SetErrorDetail
	require.Empty(t, resp.ErrorDetails)
}

func Test_ErrorDetails_NilDiffReportHandled(t *testing.T) {
	// buildContractValidationDetail should handle nil DiffReport gracefully
	cve := &ContractValidationError{
		OriginalError: fmt.Errorf("some error"),
		DiffReport:    nil,
		Scenario:      "nil-diff",
		URL:           "https://example.com",
	}
	detail := buildContractValidationDetail(cve)
	require.NotNil(t, detail)
	require.Equal(t, "nil-diff", detail.Scenario)
	require.Empty(t, detail.MissingFields)
}
