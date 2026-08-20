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

// RedactURL redacts URL user information and selected query or fragment keys.
// It falls back to conservative raw-string redaction when parsing fails.
func RedactURL(rawURL string, keys ...string) string {
	return redactURL(rawURL, querySecretKeys(keys), true)
}

// RedactURLQuery redacts selected query or fragment keys in a URL string while
// preserving any user information. Use RedactURL when a value can appear in an
// error or log.
func RedactURLQuery(rawURL string, keys ...string) string {
	return redactURL(rawURL, querySecretKeys(keys), false)
}

func redactURL(rawURL string, sensitive map[string]struct{}, redactUserInfo bool) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return redactRawURL(rawURL, sensitive, redactUserInfo)
	}

	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return redactRawURL(rawURL, sensitive, redactUserInfo)
	}
	if redactUserInfo {
		parsed.User = nil
	}
	for key := range query {
		if _, ok := sensitive[strings.ToLower(key)]; ok {
			query.Set(key, "[REDACTED]")
		}
	}
	parsed.RawQuery = query.Encode()
	parsed.Fragment = redactFragment(parsed.Fragment, sensitive)
	return parsed.String()
}

func redactFragment(fragment string, sensitive map[string]struct{}) string {
	if fragment == "" {
		return ""
	}
	values, err := url.ParseQuery(fragment)
	if err != nil {
		return redactRawParameters(fragment, sensitive)
	}
	for key := range values {
		if _, ok := sensitive[strings.ToLower(key)]; ok {
			values.Set(key, "[REDACTED]")
		}
	}
	return values.Encode()
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
		prefix, fragment, hasFragment := strings.Cut(rawURL, "#")
		if !hasFragment {
			return rawURL
		}
		return prefix + "#" + redactRawParameters(fragment, sensitive)
	}
	rawQuery, fragment, hasFragment := strings.Cut(rawQuery, "#")
	rawQuery = redactRawParameters(rawQuery, sensitive)
	redacted := prefix + "?" + rawQuery
	if hasFragment {
		redacted += "#" + redactRawParameters(fragment, sensitive)
	}
	return redacted
}

func redactRawParameters(raw string, sensitive map[string]struct{}) string {
	parts := strings.Split(raw, "&")
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
	return strings.Join(parts, "&")
}

func redactRawURL(rawURL string, sensitive map[string]struct{}, redactUserInfo bool) string {
	redacted := redactRawQuery(rawURL, sensitive)
	if !redactUserInfo {
		return redacted
	}
	return redactRawUserInfo(redacted)
}

func redactRawUserInfo(rawURL string) string {
	authorityStart := 0
	switch {
	case strings.HasPrefix(rawURL, "//"):
		authorityStart = 2
	case strings.Index(rawURL, "://") >= 0:
		authorityStart = strings.Index(rawURL, "://") + len("://")
	default:
		return rawURL
	}

	authorityEnd := len(rawURL)
	if offset := strings.IndexAny(rawURL[authorityStart:], "/?#"); offset >= 0 {
		authorityEnd = authorityStart + offset
	}
	at := strings.LastIndex(rawURL[authorityStart:authorityEnd], "@")
	if at < 0 {
		return rawURL
	}
	at += authorityStart
	return rawURL[:authorityStart] + "[REDACTED]@" + rawURL[at+1:]
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
