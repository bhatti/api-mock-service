// SPDX-License-Identifier: MIT

package contract

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
)

// ProducerExecutorFailureDetector adapts ProducerExecutor to the shrink.FailureDetector
// interface. It tests a single scenario against the remote service and reports whether it fails.
type ProducerExecutorFailureDetector struct {
	executor     *ProducerExecutor
	baseURL      string
	dataTemplate fuzz.DataTemplateRequest
	contractReq  *types.ProducerContractRequest
}

// NewProducerExecutorFailureDetector creates a failure detector backed by a ProducerExecutor.
func NewProducerExecutorFailureDetector(
	executor *ProducerExecutor,
	baseURL string,
	dataTemplate fuzz.DataTemplateRequest,
	contractReq *types.ProducerContractRequest,
) *ProducerExecutorFailureDetector {
	return &ProducerExecutorFailureDetector{
		executor:     executor,
		baseURL:      baseURL,
		dataTemplate: dataTemplate,
		contractReq:  contractReq,
	}
}

// TestScenario executes the scenario once and returns non-nil if it fails.
// Implements shrink.FailureDetector.
func (d *ProducerExecutorFailureDetector) TestScenario(ctx context.Context, scenario *types.APIScenario) error {
	key := scenario.ToKeyData()
	req := &http.Request{}
	singleReq := &types.ProducerContractRequest{
		BaseURL:        d.baseURL,
		ExecutionTimes: 1,
		Verbose:        false,
		Headers:        make(map[string][]string),
		Params:         make(map[string]any),
	}
	res := d.executor.Execute(ctx, req, key, d.dataTemplate, singleReq)
	if len(res.Errors) > 0 {
		for _, msg := range res.Errors {
			return fmt.Errorf("%s", msg)
		}
	}
	return nil
}
