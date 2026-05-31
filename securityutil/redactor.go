package securityutil

import (
	"net/url"
	"regexp"
	"strings"
)

// Pattern describes a redaction regexp and replacement.
type Pattern struct {
	Name        string
	Expr        string
	Replacement string
}

// Redactor applies configured redaction patterns to text.
type Redactor struct {
	patterns []compiledPattern
}

type compiledPattern struct {
	name        string
	expr        *regexp.Regexp
	replacement string
}

// DefaultRedactor returns a conservative redactor for common token and password forms.
func DefaultRedactor() *Redactor {
	r, err := NewRedactor(DefaultPatterns()...)
	if err != nil {
		panic(err)
	}
	return r
}

// DefaultPatterns returns the built-in conservative redaction patterns.
func DefaultPatterns() []Pattern {
	return []Pattern{
		{
			Name:        "authorization-bearer",
			Expr:        `(?i)(Authorization\s*:\s*Bearer\s+)[A-Za-z0-9._~+/=-]+`,
			Replacement: `${1}[REDACTED]`,
		},
		{
			Name:        "query-secret",
			Expr:        `(?i)([?&](?:token|secret|password|api_key)=)[^&\s]+`,
			Replacement: `${1}[REDACTED]`,
		},
		{
			Name:        "key-value-secret",
			Expr:        `(?i)\b(token|secret|password|api_key)\b(\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,;]+)`,
			Replacement: `${1}${2}[REDACTED]`,
		},
	}
}

// NewRedactor compiles patterns into a Redactor.
func NewRedactor(patterns ...Pattern) (*Redactor, error) {
	compiled := make([]compiledPattern, 0, len(patterns))
	for _, pattern := range patterns {
		expr, err := regexp.Compile(pattern.Expr)
		if err != nil {
			return nil, err
		}
		replacement := pattern.Replacement
		if replacement == "" {
			replacement = "[REDACTED]"
		}
		compiled = append(compiled, compiledPattern{
			name:        pattern.Name,
			expr:        expr,
			replacement: replacement,
		})
	}
	return &Redactor{patterns: compiled}, nil
}

// Redact applies all configured patterns to s.
func (r *Redactor) Redact(s string) string {
	if r == nil || s == "" {
		return s
	}
	out := s
	for _, pattern := range r.patterns {
		out = pattern.expr.ReplaceAllString(out, pattern.replacement)
	}
	return out
}

// RedactURLQuery redacts selected query keys in a URL string.
func RedactURLQuery(rawURL string, keys ...string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if len(keys) == 0 {
		keys = []string{"token", "secret", "password", "api_key"}
	}
	sensitive := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		sensitive[strings.ToLower(key)] = struct{}{}
	}

	query := parsed.Query()
	for key := range query {
		if _, ok := sensitive[strings.ToLower(key)]; ok {
			query.Set(key, "[REDACTED]")
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
