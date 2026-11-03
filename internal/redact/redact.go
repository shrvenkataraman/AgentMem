// Package redact masks secrets before text is stored.
package redact

// Text returns s with every secret-shaped substring replaced by "[REDACTED:<kind>]".
func Text(s string) string {
	panic("not implemented")
}
