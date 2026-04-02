package speccompare

import (
	"context"
	"testing"

	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
)


const baseSpecYAML = `
openapi: "3.0.0"
info:
  title: TestAPI
  version: "1.0.0"
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: page
          in: query
          required: false
          schema:
            type: integer
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                required: [id, name]
                properties:
                  id:
                    type: integer
                  name:
                    type: string
                  email:
                    type: string
  /orders:
    post:
      summary: Create order
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                item:
                  type: string
      responses:
        "201":
          description: Created
`

const headSpecRemovedPath = `
openapi: "3.0.0"
info:
  title: TestAPI
  version: "2.0.0"
paths:
  /users:
    get:
      summary: List users
      responses:
        "200":
          description: OK
`

const headSpecAddedPath = `
openapi: "3.0.0"
info:
  title: TestAPI
  version: "2.0.0"
paths:
  /users:
    get:
      summary: List users
      responses:
        "200":
          description: OK
  /orders:
    post:
      summary: Create order
      responses:
        "201":
          description: Created
  /products:
    get:
      summary: List products
      responses:
        "200":
          description: OK
`

const headSpecNewRequiredParam = `
openapi: "3.0.0"
info:
  title: TestAPI
  version: "2.0.0"
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: page
          in: query
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: OK
  /orders:
    post:
      summary: Create order
      responses:
        "201":
          description: Created
`

const headSpecTypeChange = `
openapi: "3.0.0"
info:
  title: TestAPI
  version: "2.0.0"
paths:
  /users:
    get:
      summary: List users
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                required: [id, name]
                properties:
                  id:
                    type: string
                  name:
                    type: string
  /orders:
    post:
      summary: Create order
      responses:
        "201":
          description: Created
`

func TestDiff_IdenticalSpecs_NoChanges(t *testing.T) {
	cfg := &types.Configuration{}
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	_, _, base, err := oapi.Parse(context.Background(), cfg, []byte(baseSpecYAML), dt)
	require.NoError(t, err)
	_, _, head, err2 := oapi.Parse(context.Background(), cfg, []byte(baseSpecYAML), dt)
	require.NoError(t, err2)

	report := Diff(base, head)
	require.False(t, report.HasBreakingChanges())
	require.Empty(t, report.BreakingChanges)
	require.Empty(t, report.RemovedPaths)
}

func TestDiff_RemovedPath_IsBreaking(t *testing.T) {
	cfg := &types.Configuration{}
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	_, _, base, _ := oapi.Parse(context.Background(), cfg, []byte(baseSpecYAML), dt)
	_, _, head, _ := oapi.Parse(context.Background(), cfg, []byte(headSpecRemovedPath), dt)

	report := Diff(base, head)
	require.True(t, report.HasBreakingChanges())
	require.Contains(t, report.RemovedPaths, "/orders")
}

func TestDiff_AddedPath_IsNonBreaking(t *testing.T) {
	cfg := &types.Configuration{}
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	_, _, base, _ := oapi.Parse(context.Background(), cfg, []byte(baseSpecYAML), dt)
	_, _, head, _ := oapi.Parse(context.Background(), cfg, []byte(headSpecAddedPath), dt)

	report := Diff(base, head)
	require.False(t, report.HasBreakingChanges())
	require.Contains(t, report.AddedPaths, "/products")
}

func TestDiff_NewRequiredParam_IsBreaking(t *testing.T) {
	cfg := &types.Configuration{}
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	_, _, base, _ := oapi.Parse(context.Background(), cfg, []byte(baseSpecYAML), dt)
	_, _, head, _ := oapi.Parse(context.Background(), cfg, []byte(headSpecNewRequiredParam), dt)

	report := Diff(base, head)
	require.True(t, report.HasBreakingChanges())
	found := false
	for _, c := range report.BreakingChanges {
		if c.ChangeType == ChangeTypeNewRequired {
			found = true
			break
		}
	}
	require.True(t, found, "expected new-required-param breaking change")
}

func TestDiff_TypeChange_IsBreaking(t *testing.T) {
	cfg := &types.Configuration{}
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	_, _, base, _ := oapi.Parse(context.Background(), cfg, []byte(baseSpecYAML), dt)
	_, _, head, _ := oapi.Parse(context.Background(), cfg, []byte(headSpecTypeChange), dt)

	report := Diff(base, head)
	require.True(t, report.HasBreakingChanges())
	found := false
	for _, c := range report.BreakingChanges {
		if c.ChangeType == ChangeTypeTypeChange && c.Field == "id" {
			found = true
			break
		}
	}
	require.True(t, found, "expected type-change breaking change for field 'id'")
}

func TestDiff_Summary_ContainsBreakingCount(t *testing.T) {
	cfg := &types.Configuration{}
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	_, _, base, _ := oapi.Parse(context.Background(), cfg, []byte(baseSpecYAML), dt)
	_, _, head, _ := oapi.Parse(context.Background(), cfg, []byte(headSpecRemovedPath), dt)

	report := Diff(base, head)
	summary := report.Summary()
	require.Contains(t, summary, "breaking")
}

func TestDiff_NilDocs_NoChange(t *testing.T) {
	report := Diff(nil, nil)
	require.False(t, report.HasBreakingChanges())
	require.Empty(t, report.BreakingChanges)
}
