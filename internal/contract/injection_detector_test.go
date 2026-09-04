package contract

import (
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDetector() *InjectionDetector {
	return NewInjectionDetector()
}

func TestInjectionDetector_SQLErrorSignatures(t *testing.T) {
	d := newTestDetector()

	cases := []struct {
		name     string
		body     string
		expected string
	}{
		{"MySQL", `{"error": "You have an error in your SQL syntax near 'x'"}`, "sqli-error-mysql"},
		{"PostgreSQL", `{"detail": "ERROR: syntax error at or near \"x\""}`, "sqli-error-postgresql"},
		{"SQLite", `{"error": "SQLITE_ERROR: no such table"}`, "sqli-error-sqlite"},
		{"MSSQL", `{"error": "Incorrect syntax near 'x'"}`, "sqli-error-mssql"},
		{"Oracle", `{"error": "ORA-00942: table or view does not exist"}`, "sqli-error-oracle"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings := d.Analyze("' OR 1=1--", "username", "POST /login", tc.body, 500)
			require.NotEmpty(t, findings)
			found := false
			for _, f := range findings {
				if f.Type == tc.expected {
					assert.Equal(t, "critical", f.Severity)
					assert.Equal(t, "CWE-89", f.CWE)
					assert.Equal(t, "username", f.Field)
					assert.Equal(t, "POST /login", f.Endpoint)
					found = true
				}
			}
			assert.True(t, found, "expected finding type %s", tc.expected)
		})
	}
}

func TestInjectionDetector_NoSQLErrorSignatures(t *testing.T) {
	d := newTestDetector()
	body := `{"error": "MongoError: bad auth Authentication failed"}`
	findings := d.Analyze(`{"$gt": ""}`, "filter", "GET /users", body, 500)
	require.NotEmpty(t, findings)
	hasFinding := false
	for _, f := range findings {
		if f.CWE == "CWE-943" {
			hasFinding = true
		}
	}
	assert.True(t, hasFinding)
}

func TestInjectionDetector_LDAPErrorSignatures(t *testing.T) {
	d := newTestDetector()
	body := `javax.naming.NameNotFoundException: cn=admin,dc=example`
	findings := d.Analyze("*(|(uid=*))", "query", "GET /search", body, 500)
	require.NotEmpty(t, findings)
	hasFinding := false
	for _, f := range findings {
		if f.CWE == "CWE-90" {
			hasFinding = true
		}
	}
	assert.True(t, hasFinding)
}

func TestInjectionDetector_ReflectedXSS(t *testing.T) {
	d := newTestDetector()
	payload := `<script>alert(1)</script>`
	body := `<html><body>Hello <script>alert(1)</script></body></html>`
	findings := d.Analyze(payload, "name", "GET /profile", body, 200)
	require.NotEmpty(t, findings)
	found := false
	for _, f := range findings {
		if f.Type == "xss-reflected" {
			assert.Equal(t, "high", f.Severity)
			assert.Equal(t, "CWE-79", f.CWE)
			found = true
		}
	}
	assert.True(t, found)
}

func TestInjectionDetector_ReflectedSSTI(t *testing.T) {
	d := newTestDetector()
	payload := "${7*7}"
	body := `{"result": "${7*7}"}`
	findings := d.Analyze(payload, "template", "POST /render", body, 200)
	require.NotEmpty(t, findings)
	found := false
	for _, f := range findings {
		if f.Type == "ssti-reflected" {
			assert.Equal(t, "CWE-1336", f.CWE)
			found = true
		}
	}
	assert.True(t, found)
}

func TestInjectionDetector_Unexpected500(t *testing.T) {
	d := newTestDetector()
	findings := d.Analyze("' OR 1=1--", "input", "POST /api", `{"error": "internal"}`, 500)
	found := false
	for _, f := range findings {
		if f.Type == "unhandled-input" {
			assert.Equal(t, "medium", f.Severity)
			assert.Equal(t, "CWE-20", f.CWE)
			found = true
		}
	}
	assert.True(t, found)
}

func TestInjectionDetector_CleanResponse(t *testing.T) {
	d := newTestDetector()
	findings := d.Analyze("' OR 1=1--", "input", "POST /api", `{"status": "ok", "data": [1,2,3]}`, 200)
	assert.Empty(t, findings)
}

func TestInjectionDetector_EmptyResponse(t *testing.T) {
	d := newTestDetector()
	findings := d.Analyze("test", "field", "GET /api", "", 200)
	assert.Empty(t, findings)
}

func TestInjectionDetector_InfoDisclosure_StackTrace(t *testing.T) {
	d := newTestDetector()
	body := `Exception in thread "main" at com.example.App.java:42`
	findings := d.Analyze("test", "field", "GET /api", body, 500)
	hasInfo := false
	for _, f := range findings {
		if f.Severity == "info" && f.CWE == "CWE-200" {
			hasInfo = true
		}
	}
	assert.True(t, hasInfo)
}

func TestInjectionDetector_InfoDisclosure_FilePath(t *testing.T) {
	d := newTestDetector()
	body := `{"error": "File not found: /var/log/app.log"}`
	findings := d.Analyze("test", "field", "GET /api", body, 404)
	hasInfo := false
	for _, f := range findings {
		if f.Severity == "info" && f.CWE == "CWE-200" {
			hasInfo = true
		}
	}
	assert.True(t, hasInfo)
}

// Timing-based detection

func TestInjectionDetector_TimingBlindInjection(t *testing.T) {
	d := newTestDetector()
	findings := d.AnalyzeWithTiming(
		"' OR SLEEP(5)--", "input", "POST /login",
		`{"status": "ok"}`, 200,
		15000, 2000, 3.0,
	)
	found := false
	for _, f := range findings {
		if f.Type == "blind-timing" {
			assert.Equal(t, "critical", f.Severity)
			assert.Equal(t, "CWE-89", f.CWE)
			assert.Contains(t, f.Evidence, "15000ms")
			found = true
		}
	}
	assert.True(t, found)
}

func TestInjectionDetector_TimingWithoutKeyword_NotFlagged(t *testing.T) {
	d := newTestDetector()
	findings := d.AnalyzeWithTiming(
		"normal input", "input", "POST /api",
		`{"status": "ok"}`, 200,
		15000, 2000, 3.0,
	)
	for _, f := range findings {
		assert.NotEqual(t, "blind-timing", f.Type, "should not flag timing without timing keyword")
	}
}

func TestInjectionDetector_TimingBelowThreshold_NotFlagged(t *testing.T) {
	d := newTestDetector()
	findings := d.AnalyzeWithTiming(
		"' OR SLEEP(5)--", "input", "POST /login",
		`{"status": "ok"}`, 200,
		3000, 2000, 3.0,
	)
	for _, f := range findings {
		assert.NotEqual(t, "blind-timing", f.Type, "should not flag when below threshold")
	}
}

func TestInjectionDetector_TimingDefaultThreshold(t *testing.T) {
	d := newTestDetector()
	findings := d.AnalyzeWithTiming(
		"' OR SLEEP(5)--", "input", "POST /login",
		`{"status": "ok"}`, 200,
		15000, 2000, 0, // 0 → use default 3.0
	)
	found := false
	for _, f := range findings {
		if f.Type == "blind-timing" {
			found = true
		}
	}
	assert.True(t, found)
}

// Security Summary

func TestInjectionDetector_Summarize(t *testing.T) {
	d := newTestDetector()
	findings := []types.InjectionFinding{
		{Type: "sqli-error-mysql", Severity: "critical", CWE: "CWE-89"},
		{Type: "sqli-error-mysql", Severity: "critical", CWE: "CWE-89"},

		{Type: "sqli-error-mysql", Severity: "critical", CWE: "CWE-89"},

		{Type: "xss-reflected", Severity: "high", CWE: "CWE-79"},
		{Type: "xss-reflected", Severity: "high", CWE: "CWE-79"},
		{Type: "unhandled-input", Severity: "medium", CWE: "CWE-20"},
	}

	summary := d.Summarize(findings)
	assert.Equal(t, 6, summary.TotalFindings)
	assert.Equal(t, 3, summary.BySeverity["critical"])
	assert.Equal(t, 2, summary.BySeverity["high"])
	assert.Equal(t, 1, summary.BySeverity["medium"])
	assert.Equal(t, 3, summary.ByCategory["CWE-89"])
	assert.Equal(t, 2, summary.ByCategory["CWE-79"])
	assert.Equal(t, 6, len(summary.TopFindings))
	assert.Equal(t, "critical", summary.TopFindings[0].Severity)
	assert.NotEmpty(t, summary.OWASPMapping)
	assert.Contains(t, summary.OWASPMapping, "CWE-89")
}

func TestInjectionDetector_Summarize_NoFindings(t *testing.T) {
	d := newTestDetector()
	summary := d.Summarize(nil)
	assert.Equal(t, 0, summary.TotalFindings)
	assert.Equal(t, len(d.allCheckClasses), len(summary.PassedChecks))
}

func TestInjectionDetector_Summarize_AllCWEsMapped(t *testing.T) {
	mapping := owaspMapping()
	expectedCWEs := []string{"CWE-89", "CWE-79", "CWE-78", "CWE-22", "CWE-90", "CWE-611", "CWE-1336", "CWE-943"}
	for _, cwe := range expectedCWEs {
		assert.Contains(t, mapping, cwe, "OWASP mapping missing %s", cwe)
	}
}

func TestSortBySeverity(t *testing.T) {
	findings := []types.InjectionFinding{
		{Type: "a", Severity: "info"},
		{Type: "b", Severity: "critical"},
		{Type: "c", Severity: "medium"},
		{Type: "d", Severity: "high"},
	}
	sorted := sortBySeverity(findings)
	assert.Equal(t, "critical", sorted[0].Severity)
	assert.Equal(t, "high", sorted[1].Severity)
	assert.Equal(t, "medium", sorted[2].Severity)
	assert.Equal(t, "info", sorted[3].Severity)
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "abc", truncate("abc", 10))
	assert.Equal(t, "abcde...", truncate("abcdefghij", 5))
}
