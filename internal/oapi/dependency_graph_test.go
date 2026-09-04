package oapi

import (
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func petStoreSpecs() []*APISpec {
	return []*APISpec{
		{
			Method: types.Post,
			Path:   "/pets",
			Request: Request{
				Body: []Property{
					{Name: "name", Type: "string"},
					{Name: "tag", Type: "string"},
				},
			},
			Response: Response{
				StatusCode: 201,
				Body: []Property{
					{Name: "id", Type: "integer"},
					{Name: "name", Type: "string"},
				},
			},
		},
		{
			Method: types.Get,
			Path:   "/pets/{petId}",
			Request: Request{
				PathParams: []Property{{Name: "petId", Type: "integer"}},
			},
			Response: Response{
				StatusCode: 200,
				Body: []Property{
					{Name: "id", Type: "integer"},
					{Name: "name", Type: "string"},
				},
			},
		},
		{
			Method: types.Put,
			Path:   "/pets/{petId}",
			Request: Request{
				PathParams: []Property{{Name: "petId", Type: "integer"}},
				Body:       []Property{{Name: "name", Type: "string"}},
			},
			Response: Response{
				StatusCode: 200,
				Body: []Property{
					{Name: "id", Type: "integer"},
					{Name: "name", Type: "string"},
				},
			},
		},
		{
			Method: types.Delete,
			Path:   "/pets/{petId}",
			Request: Request{
				PathParams: []Property{{Name: "petId", Type: "integer"}},
			},
			Response: Response{StatusCode: 204},
		},
	}
}

func TestBuildNodes_PetStore(t *testing.T) {
	specs := petStoreSpecs()
	nodes := buildNodes(specs)
	require.Len(t, nodes, 4)

	postNode := nodes[0]
	assert.Equal(t, "POST", postNode.Method)
	assert.Contains(t, postNode.Produces, "id")
	assert.Contains(t, postNode.Produces, "name")

	getNode := nodes[1]
	assert.Equal(t, "GET", getNode.Method)
	assert.Contains(t, getNode.Consumes, "petid")
}

func TestFindEdges_PetStore(t *testing.T) {
	specs := petStoreSpecs()
	nodes := buildNodes(specs)
	edges := findEdges(nodes)
	t.Logf("edges: %d", len(edges))
	for _, e := range edges {
		t.Logf("  %s %s → %s %s (field=%s, confidence=%.1f)", e.From.Method, e.From.Path, e.To.Method, e.To.Path, e.FieldName, e.Confidence)
	}
	require.NotEmpty(t, edges)
}

func TestBuildDependencyGraph_PetStore(t *testing.T) {
	specs := petStoreSpecs()
	edges, groups := BuildDependencyGraph(specs)

	require.NotEmpty(t, edges)
	require.NotEmpty(t, groups)

	hasPostToGet := false
	for _, e := range edges {
		if e.From.Method == "POST" && e.From.Path == "/pets" &&
			e.To.Path == "/pets/{petId}" {
			hasPostToGet = true
			assert.True(t, e.Confidence > 0)
		}
	}
	assert.True(t, hasPostToGet, "expected POST /pets → GET/PUT/DELETE /pets/{petId} edge")
}

func TestBuildDependencyGraph_NoMatchingFields(t *testing.T) {
	specs := []*APISpec{
		{
			Method: types.Get,
			Path:   "/alpha",
			Response: Response{
				Body: []Property{{Name: "foo", Type: "string"}},
			},
		},
		{
			Method: types.Get,
			Path:   "/beta",
			Request: Request{
				QueryParams: []Property{{Name: "bar", Type: "string"}},
			},
			Response: Response{
				Body: []Property{{Name: "baz", Type: "string"}},
			},
		},
	}

	edges, groups := BuildDependencyGraph(specs)
	assert.Empty(t, edges)
	require.NotEmpty(t, groups)
	assert.Equal(t, 2, len(groups[0]))
}

func TestBuildDependencyGraph_TopologicalOrder(t *testing.T) {
	specs := petStoreSpecs()
	_, groups := BuildDependencyGraph(specs)
	require.NotEmpty(t, groups, "expected at least one topological group")

	totalNodes := 0
	for _, g := range groups {
		totalNodes += len(g)
	}
	assert.Equal(t, 4, totalNodes, "all 4 pet store operations should appear in groups")
}

func TestApplyDependencies(t *testing.T) {
	specs := petStoreSpecs()
	edges, _ := BuildDependencyGraph(specs)
	ApplyDependencies(specs, edges)

	hasOrder := false
	for _, spec := range specs {
		if spec.DependencyOrder >= 0 {
			hasOrder = true
		}
	}
	assert.True(t, hasOrder, "at least one spec should have a dependency order assigned")
}

func TestApplyDependencies_LinearChain(t *testing.T) {
	specs := []*APISpec{
		{
			Method: types.Post, Path: "/orders",
			Response: Response{Body: []Property{{Name: "orderId", Type: "string"}}},
		},
		{
			Method: types.Get, Path: "/orders/{orderId}",
			Request: Request{PathParams: []Property{{Name: "orderId", Type: "string"}}},
			Response: Response{Body: []Property{{Name: "orderId", Type: "string"}, {Name: "status", Type: "string"}}},
		},
	}
	edges, _ := BuildDependencyGraph(specs)
	ApplyDependencies(specs, edges)

	orderMap := make(map[string]int)
	for _, spec := range specs {
		key := string(spec.Method) + " " + spec.Path
		orderMap[key] = spec.DependencyOrder
	}
	assert.LessOrEqual(t, orderMap["POST /orders"], orderMap["GET /orders/{orderId}"])
}

func TestNormalizeName(t *testing.T) {
	assert.Equal(t, "petid", normalizeName("pet_id"))
	assert.Equal(t, "petid", normalizeName("petId"))
	assert.Equal(t, "petid", normalizeName("pet-id"))
	assert.Equal(t, "userid", normalizeName("userId"))
}

func TestSingularPlural(t *testing.T) {
	assert.True(t, singularPlural("pet", "pets"))
	assert.True(t, singularPlural("pets", "pet"))
	assert.True(t, singularPlural("address", "addresses"))
	assert.False(t, singularPlural("cat", "dog"))
}

func TestExtractResource(t *testing.T) {
	assert.Equal(t, "pets", extractResource("/pets/{petId}"))
	assert.Equal(t, "users", extractResource("/api/v1/users/{userId}"))
	assert.Equal(t, "items", extractResource("/items"))
}

func TestMatchConfidence_Exact(t *testing.T) {
	assert.Equal(t, 1.0, matchConfidence("name", "name", "/a", "/b"))
	assert.Equal(t, 1.0, matchConfidence("user_id", "userId", "/a", "/b"))
}

func TestMatchConfidence_SingularPlural(t *testing.T) {
	c := matchConfidence("pet", "pets", "/a", "/b")
	assert.Equal(t, 0.9, c)
}

func TestMatchConfidence_NoMatch(t *testing.T) {
	assert.Equal(t, 0.0, matchConfidence("foo", "bar", "/a", "/b"))
}

func TestBuildDependencyGraph_Empty(t *testing.T) {
	edges, groups := BuildDependencyGraph(nil)
	assert.Empty(t, edges)
	assert.Empty(t, groups)
}

func TestBuildDependencyGraph_SingleSpec(t *testing.T) {
	specs := []*APISpec{{
		Method: types.Get,
		Path:   "/health",
		Response: Response{
			Body: []Property{{Name: "status", Type: "string"}},
		},
	}}
	edges, groups := BuildDependencyGraph(specs)
	assert.Empty(t, edges)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0], 1)
}
