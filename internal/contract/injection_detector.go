package contract

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bhatti/api-mock-service/internal/types"
)

// compiledSignature holds a pre-compiled regex with its metadata.
type compiledSignature struct {
	re       *regexp.Regexp
	findType string
	severity string
	cwe      string
}

// InjectionDetector scans HTTP responses for evidence of injection vulnerabilities.
// It pre-compiles all signature patterns at construction time for efficient reuse.
type InjectionDetector struct {
	errorSignatures []compiledSignature
	infoSignatures  []compiledSignature
	timingKeywords  []string
	allCheckClasses []string
}

// NewInjectionDetector creates a detector with all signature patterns pre-compiled.
func NewInjectionDetector() *InjectionDetector {
	d := &InjectionDetector{
		timingKeywords: []string{
			"SLEEP", "WAITFOR", "pg_sleep", "BENCHMARK", "sleep",
		},
		allCheckClasses: []string{
			"sqli", "xss", "path-traversal", "ssti", "cmd-injection",
			"nosqli", "ldapi", "xxe", "info-disclosure",
		},
	}

	// SQL error signatures
	sqlPatterns := []struct {
		pattern string
		db      string
	}{
		{`(?i)you have an error in your sql syntax`, "mysql"},
		{`(?i)warning:.*mysql`, "mysql"},
		{`(?i)unclosed quotation mark after the character string`, "mssql"},
		{`(?i)incorrect syntax near`, "mssql"},
		{`(?i)error.*syntax error at or near`, "postgresql"},
		{`(?i)pg_query\(\):.*error`, "postgresql"},
		{`(?i)SQLITE_ERROR`, "sqlite"},
		{`(?i)sqlite3\.OperationalError`, "sqlite"},
		{`(?i)ORA-\d{5}`, "oracle"},
		{`(?i)quoted string not properly terminated`, "oracle"},
		{`(?i)com\.mysql\.jdbc`, "mysql"},
		{`(?i)org\.postgresql\.util\.PSQLException`, "postgresql"},
		{`(?i)System\.Data\.SqlClient\.SqlException`, "mssql"},
		{`(?i)java\.sql\.SQLException`, "java-sql"},
	}
	for _, sp := range sqlPatterns {
		re := regexp.MustCompile(sp.pattern)
		d.errorSignatures = append(d.errorSignatures, compiledSignature{
			re: re, findType: "sqli-error-" + sp.db, severity: "critical", cwe: "CWE-89",
		})
	}

	// NoSQL error signatures
	nosqlPatterns := []string{
		`(?i)MongoError`,
		`(?i)MongoDB.*Error`,
		`(?i)com\.mongodb\.MongoException`,
		`(?i)CouchDB.*error`,
		`(?i)cassandra\.InvalidRequest`,
	}
	for _, p := range nosqlPatterns {
		re := regexp.MustCompile(p)
		d.errorSignatures = append(d.errorSignatures, compiledSignature{
			re: re, findType: "nosqli-error", severity: "critical", cwe: "CWE-943",
		})
	}

	// LDAP error signatures
	ldapPatterns := []string{
		`(?i)javax\.naming\.NameNotFoundException`,
		`(?i)javax\.naming\.directory`,
		`(?i)LDAPException`,
		`(?i)InvalidSearchFilterException`,
		`(?i)LDAP.*error`,
	}
	for _, p := range ldapPatterns {
		re := regexp.MustCompile(p)
		d.errorSignatures = append(d.errorSignatures, compiledSignature{
			re: re, findType: "ldapi-error", severity: "high", cwe: "CWE-90",
		})
	}

	// Information disclosure signatures
	infoPatterns := []struct {
		pattern  string
		findType string
	}{
		{`(?i)at [a-zA-Z0-9_.]+\.(java|py|rb|cs|go|php|js|ts):\d+`, "stack-trace"},
		{`(?i)traceback \(most recent call last\)`, "stack-trace-python"},
		{`(?i)Fatal error:.*on line \d+`, "stack-trace-php"},
		{`/usr/[a-z]+/`, "file-path-unix"},
		{`/var/[a-z]+/`, "file-path-unix"},
		{`/home/[a-z]+/`, "file-path-unix"},
		{`(?i)[A-Z]:\\(Windows|Users|Program Files)`, "file-path-windows"},
		{`(?i)(apache|nginx|tomcat|iis)/[\d.]+`, "server-version"},
	}
	for _, ip := range infoPatterns {
		re := regexp.MustCompile(ip.pattern)
		d.infoSignatures = append(d.infoSignatures, compiledSignature{
			re: re, findType: "info-disclosure-" + ip.findType, severity: "info", cwe: "CWE-200",
		})
	}

	return d
}

// Analyze checks the response body for injection signatures and reflected payloads.
func (d *InjectionDetector) Analyze(payload, field, endpoint, responseBody string, statusCode int) []types.InjectionFinding {
	if responseBody == "" {
		return nil
	}

	var findings []types.InjectionFinding

	for _, sig := range d.errorSignatures {
		loc := sig.re.FindStringIndex(responseBody)
		if loc != nil {
			findings = append(findings, types.InjectionFinding{
				Type:     sig.findType,
				Severity: sig.severity,
				Evidence: truncate(responseBody[loc[0]:loc[1]], 200),
				Pattern:  sig.re.String(),
				Payload:  payload,
				Field:    field,
				Endpoint: endpoint,
				CWE:      sig.cwe,
			})
		}
	}

	if payload != "" && len(payload) >= 3 {
		if strings.Contains(responseBody, payload) {
			findings = append(findings, types.InjectionFinding{
				Type:     classifyReflectedPayload(payload),
				Severity: "high",
				Evidence: truncate(payload, 200),
				Pattern:  "reflected-payload",
				Payload:  payload,
				Field:    field,
				Endpoint: endpoint,
				CWE:      classifyReflectedCWE(payload),
			})
		}
	}

	if statusCode >= 500 && payload != "" {
		findings = append(findings, types.InjectionFinding{
			Type:     "unhandled-input",
			Severity: "medium",
			Evidence: fmt.Sprintf("status %d after injection payload", statusCode),
			Pattern:  "status-5xx",
			Payload:  payload,
			Field:    field,
			Endpoint: endpoint,
			CWE:      "CWE-20",
		})
	}

	for _, sig := range d.infoSignatures {
		loc := sig.re.FindStringIndex(responseBody)
		if loc != nil {
			findings = append(findings, types.InjectionFinding{
				Type:     sig.findType,
				Severity: sig.severity,
				Evidence: truncate(responseBody[loc[0]:loc[1]], 200),
				Pattern:  sig.re.String(),
				Payload:  payload,
				Field:    field,
				Endpoint: endpoint,
				CWE:      sig.cwe,
			})
		}
	}

	return findings
}

// AnalyzeWithTiming extends Analyze with timing-based blind injection detection.
func (d *InjectionDetector) AnalyzeWithTiming(
	payload, field, endpoint, responseBody string,
	statusCode int, elapsedMs int64, baselineMs int64, thresholdMultiplier float64,
) []types.InjectionFinding {
	findings := d.Analyze(payload, field, endpoint, responseBody, statusCode)

	if thresholdMultiplier <= 0 {
		thresholdMultiplier = 3.0
	}
	if baselineMs > 0 && elapsedMs > int64(float64(baselineMs)*thresholdMultiplier) {
		if d.containsTimingKeyword(payload) {
			findings = append(findings, types.InjectionFinding{
				Type:     "blind-timing",
				Severity: "critical",
				Evidence: fmt.Sprintf("elapsed %dms vs baseline %dms (%.1fx)", elapsedMs, baselineMs, float64(elapsedMs)/float64(baselineMs)),
				Pattern:  "timing-anomaly",
				Payload:  payload,
				Field:    field,
				Endpoint: endpoint,
				CWE:      "CWE-89",
			})
		}
	}

	return findings
}

// Summarize aggregates a list of findings into a SecuritySummary with OWASP classification.
func (d *InjectionDetector) Summarize(findings []types.InjectionFinding) types.SecuritySummary {
	summary := types.SecuritySummary{
		TotalFindings: len(findings),
		BySeverity:    make(map[string]int),
		ByCategory:    make(map[string]int),
		OWASPMapping:  owaspMapping(),
	}

	seenCategories := make(map[string]bool)
	for _, f := range findings {
		summary.BySeverity[f.Severity]++
		summary.ByCategory[f.CWE]++
		seenCategories[classifyCheckClass(f.Type)] = true
	}

	limit := len(findings)
	if limit > 10 {
		limit = 10
	}
	sorted := sortBySeverity(findings)
	summary.TopFindings = sorted[:limit]

	for _, cls := range d.allCheckClasses {
		if !seenCategories[cls] {
			summary.PassedChecks = append(summary.PassedChecks, cls)
		}
	}

	return summary
}

func (d *InjectionDetector) containsTimingKeyword(payload string) bool {
	upper := strings.ToUpper(payload)
	for _, kw := range d.timingKeywords {
		if strings.Contains(upper, strings.ToUpper(kw)) {
			return true
		}
	}
	return false
}

func owaspMapping() map[string]string {
	return map[string]string{
		"CWE-89":   "WSTG-INPV-05",
		"CWE-79":   "WSTG-INPV-01",
		"CWE-78":   "WSTG-INPV-12",
		"CWE-22":   "WSTG-ATHZ-01",
		"CWE-90":   "WSTG-INPV-06",
		"CWE-611":  "WSTG-INPV-07",
		"CWE-1336": "WSTG-INPV-18",
		"CWE-943":  "WSTG-INPV-05",
		"CWE-20":   "WSTG-INPV-01",
		"CWE-200":  "WSTG-INFO-02",
	}
}

func classifyReflectedPayload(payload string) string {
	lower := strings.ToLower(payload)
	switch {
	case strings.Contains(lower, "<script") || strings.Contains(lower, "alert(") || strings.Contains(lower, "onerror"):
		return "xss-reflected"
	case strings.Contains(lower, "{{") || strings.Contains(lower, "${") || strings.Contains(lower, "<%="):
		return "ssti-reflected"
	case strings.Contains(lower, ";") && (strings.Contains(lower, "cat ") || strings.Contains(lower, "whoami") || strings.Contains(lower, "id")):
		return "cmd-injection-reflected"
	default:
		return "payload-reflected"
	}
}

func classifyReflectedCWE(payload string) string {
	lower := strings.ToLower(payload)
	switch {
	case strings.Contains(lower, "<script") || strings.Contains(lower, "alert(") || strings.Contains(lower, "onerror"):
		return "CWE-79"
	case strings.Contains(lower, "{{") || strings.Contains(lower, "${") || strings.Contains(lower, "<%="):
		return "CWE-1336"
	case strings.Contains(lower, ";") && (strings.Contains(lower, "cat ") || strings.Contains(lower, "whoami")):
		return "CWE-78"
	default:
		return "CWE-20"
	}
}

func classifyCheckClass(findingType string) string {
	switch {
	case strings.HasPrefix(findingType, "sqli") || findingType == "blind-timing":
		return "sqli"
	case strings.HasPrefix(findingType, "xss"):
		return "xss"
	case strings.HasPrefix(findingType, "ssti"):
		return "ssti"
	case strings.HasPrefix(findingType, "cmd"):
		return "cmd-injection"
	case strings.HasPrefix(findingType, "nosqli"):
		return "nosqli"
	case strings.HasPrefix(findingType, "ldapi"):
		return "ldapi"
	case strings.HasPrefix(findingType, "path"):
		return "path-traversal"
	case strings.HasPrefix(findingType, "xxe"):
		return "xxe"
	case strings.HasPrefix(findingType, "info"):
		return "info-disclosure"
	default:
		return findingType
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func sortBySeverity(findings []types.InjectionFinding) []types.InjectionFinding {
	severityOrder := map[string]int{"critical": 0, "high": 1, "medium": 2, "info": 3}
	sorted := make([]types.InjectionFinding, len(findings))
	copy(sorted, findings)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0; j-- {
			oj := severityOrder[sorted[j].Severity]
			ojm1 := severityOrder[sorted[j-1].Severity]
			if oj < ojm1 {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			}
		}
	}
	return sorted
}
