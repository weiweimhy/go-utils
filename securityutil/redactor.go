package securityutil

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var defaultQuerySecretKeys = []string{
	"access_token",
	"refresh_token",
	"client_secret",
	"token",
	"secret",
	"password",
	"api_key",
	"code",
}

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
			Expr:        `(?i)([?&](?:access_token|refresh_token|client_secret|token|secret|password|api_key|code)=)[^&\s]+`,
			Replacement: `${1}[REDACTED]`,
		},
		{
			Name:        "key-value-secret",
			Expr:        `(?i)\b(access_token|refresh_token|client_secret|token|secret|password|api_key|code)\b(\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,;]+)`,
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
	sensitive := querySecretKeys(keys)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return redactRawQuery(rawURL, sensitive)
	}

	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return redactRawQuery(rawURL, sensitive)
	}
	for key := range query {
		if _, ok := sensitive[strings.ToLower(key)]; ok {
			query.Set(key, "[REDACTED]")
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func querySecretKeys(keys []string) map[string]struct{} {
	if len(keys) == 0 {
		keys = defaultQuerySecretKeys
	}
	sensitive := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		sensitive[strings.ToLower(key)] = struct{}{}
	}
	return sensitive
}

func redactRawQuery(rawURL string, sensitive map[string]struct{}) string {
	prefix, rawQuery, found := strings.Cut(rawURL, "?")
	if !found {
		return rawURL
	}
	rawQuery, fragment, hasFragment := strings.Cut(rawQuery, "#")
	parts := strings.Split(rawQuery, "&")
	for i, part := range parts {
		key, _, hasValue := strings.Cut(part, "=")
		if !hasValue {
			continue
		}
		decodedKey, err := url.QueryUnescape(key)
		if err != nil {
			decodedKey = key
		}
		if _, ok := sensitive[strings.ToLower(decodedKey)]; ok {
			parts[i] = key + "=[REDACTED]"
		}
	}
	redacted := prefix + "?" + strings.Join(parts, "&")
	if hasFragment {
		redacted += "#" + fragment
	}
	return redacted
}

// RedactLiterals replaces the supplied non-empty secret values wherever they
// occur in s. Longer values are replaced first so overlapping values cannot
// reveal a suffix of a secret.
func RedactLiterals(s string, values ...string) string {
	if s == "" || len(values) == 0 {
		return s
	}
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			unique[value] = struct{}{}
		}
	}
	ordered := make([]string, 0, len(unique))
	for value := range unique {
		ordered = append(ordered, value)
	}
	sort.Slice(ordered, func(i, j int) bool {
		return len(ordered[i]) > len(ordered[j])
	})
	for _, value := range ordered {
		s = strings.ReplaceAll(s, value, "[REDACTED]")
	}
	return s
}
