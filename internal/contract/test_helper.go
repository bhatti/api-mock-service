package contract

import (
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"time"
)

func saveTestScenario(
	name string,
	scenarioRepo repository.APIScenarioRepository,
) (*types.APIScenario, error) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	scenario := types.APIScenario{}
	// AND valid template for random data
	err = yaml.Unmarshal(b, &scenario)
	if err != nil {
		return nil, err
	}
	err = scenarioRepo.Save(&scenario)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse("http://localhost:8080")
	if err != nil {
		return nil, err
	}
	err = scenarioRepo.SaveHistory(&scenario, u.String(), time.Now(), time.Now().Add(time.Second))
	if err != nil {
		return nil, err
	}
	return &scenario, nil
}

func getFirstMap(m map[string]any) map[string]any {
	for _, v := range m {
		if resp, ok := v.(map[string]any); ok {
			return resp
		} else if strResp, ok := v.(map[string]string); ok {
			resp := make(map[string]any)
			for kk, vv := range strResp {
				resp[kk] = vv
			}
			return resp
		}
	}
	return m
}
