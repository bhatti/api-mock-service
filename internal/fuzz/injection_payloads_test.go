package fuzz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandFromGrammar_NonEmpty(t *testing.T) {
	grammar := []string{"payload-a", "payload-b", "payload-c"}
	result := randFromGrammar(grammar)
	assert.NotEmpty(t, result)
	assert.Contains(t, grammar, result)
}

func TestRandFromGrammar_Seeded_Deterministic(t *testing.T) {
	grammar := []string{"a", "b", "c", "d", "e"}
	r1 := randFromGrammar(grammar, 42)
	r2 := randFromGrammar(grammar, 42)
	assert.Equal(t, r1, r2)
}

func TestRandFromGrammar_TokenReplacement(t *testing.T) {
	grammar := []string{"SELECT * FROM {{TBL}} WHERE {{COL}}={{INT}}"}
	result := randFromGrammar(grammar, 99)
	assert.NotContains(t, result, "{{TBL}}")
	assert.NotContains(t, result, "{{COL}}")
	assert.NotContains(t, result, "{{INT}}")
}

// generatorTestCase defines a common test case for all injection generators.
type generatorTestCase struct {
	name       string
	randFn     func(seed ...int64) string
	seededFn   func(seed int64) string
	markers    []string // at least one of these should appear in output
	minUnique  int      // minimum unique values from 100 random calls
}

func TestInjectionGenerators(t *testing.T) {
	cases := []generatorTestCase{
		{
			name:      "RandSQLi",
			randFn:    RandSQLi,
			seededFn:  SeededRandSQLi,
			markers:   []string{"'", "--", "OR", "UNION", "SELECT", "SLEEP", "DROP", "WAITFOR"},
			minUnique: 10,
		},
		{
			name:      "RandXSS",
			randFn:    RandXSS,
			seededFn:  SeededRandXSS,
			markers:   []string{"<", "alert", "script", "onerror", "onload", "javascript", "onfocus", "onmouseover", "onclick", "svg", "img"},
			minUnique: 8,
		},
		{
			name:      "RandPathTraversal",
			randFn:    RandPathTraversal,
			seededFn:  SeededRandPathTraversal,
			markers:   []string{"..", "etc", "passwd", "windows", "%2f", "%2e"},
			minUnique: 8,
		},
		{
			name:      "RandSSTI",
			randFn:    RandSSTI,
			seededFn:  SeededRandSSTI,
			markers:   []string{"7*7", "${", "#{", "<%=", "system", "freemarker", "config"},
			minUnique: 5,
		},
		{
			name:      "RandCmdInjection",
			randFn:    RandCmdInjection,
			seededFn:  SeededRandCmdInjection,
			markers:   []string{";", "|", "&", "`", "$(", "id", "whoami", "cat", "ls"},
			minUnique: 8,
		},
		{
			name:      "RandNoSQLi",
			randFn:    RandNoSQLi,
			seededFn:  SeededRandNoSQLi,
			markers:   []string{"$gt", "$ne", "$regex", "$where", "$or", "return", "sleep"},
			minUnique: 5,
		},
		{
			name:      "RandLDAPi",
			randFn:    RandLDAPi,
			seededFn:  SeededRandLDAPi,
			markers:   []string{"*", "(", ")", "uid", "cn", "objectClass"},
			minUnique: 5,
		},
		{
			name:      "RandXXE",
			randFn:    RandXXE,
			seededFn:  SeededRandXXE,
			markers:   []string{"<!DOCTYPE", "<!ENTITY", "SYSTEM", "xxe", "xml"},
			minUnique: 5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"_NonEmpty", func(t *testing.T) {
			result := tc.randFn()
			assert.NotEmpty(t, result)
		})

		t.Run(tc.name+"_Seeded_Deterministic", func(t *testing.T) {
			r1 := tc.seededFn(12345)
			r2 := tc.seededFn(12345)
			assert.Equal(t, r1, r2, "seeded calls with same seed should produce identical output")
		})

		t.Run(tc.name+"_ContainsMarker", func(t *testing.T) {
			found := false
			for i := 0; i < 20; i++ {
				result := tc.randFn()
				for _, marker := range tc.markers {
					if containsCI(result, marker) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			assert.True(t, found, "expected at least one marker %v in 20 random outputs", tc.markers)
		})

		t.Run(tc.name+"_Diversity", func(t *testing.T) {
			seen := make(map[string]bool)
			for i := 0; i < 100; i++ {
				seen[tc.randFn()] = true
			}
			assert.GreaterOrEqual(t, len(seen), tc.minUnique,
				"expected >= %d unique values from 100 calls, got %d", tc.minUnique, len(seen))
		})
	}
}

func TestAllInjectionPayloads(t *testing.T) {
	payloads := AllInjectionPayloads()
	require.Len(t, payloads, 8)
	for class, payload := range payloads {
		assert.NotEmpty(t, payload, "payload for %s should not be empty", class)
	}
}

func TestAllInjectionPayloads_Seeded_Deterministic(t *testing.T) {
	p1 := AllInjectionPayloads(42)
	p2 := AllInjectionPayloads(42)
	assert.Equal(t, p1, p2)
}

// containsCI checks if s contains substr (case-insensitive).
func containsCI(s, substr string) bool {
	return len(s) >= len(substr) &&
		(indexOf(toLower(s), toLower(substr)) >= 0)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
