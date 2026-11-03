// Package redact masks secrets before text is stored.
package redact

import "regexp"

type rule struct {
	re   *regexp.Regexp
	want string
}

var rules = []rule{
	{regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`), "[REDACTED:aws]"},
	{regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{36}\b`), "[REDACTED:github]"},
	{regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`), "[REDACTED:jwt]"},
	{regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`), "[REDACTED:private-key]"},
	{regexp.MustCompile(`(?i)\b([A-Z0-9_]*(?:TOKEN|SECRET|PASSWORD|PASSWD|API_?KEY|PRIVATE_?KEY)[A-Z0-9_]*)=("[^"]*"|'[^']*'|[^\s'"]+)`), "${1}=[REDACTED:secret]"},
}

// Text returns s with every secret-shaped substring replaced by "[REDACTED:<kind>]".
func Text(s string) string {
	for _, r := range rules {
		s = r.re.ReplaceAllString(s, r.want)
	}
	return s
}
