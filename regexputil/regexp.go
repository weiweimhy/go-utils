package regexputil

import (
	"regexp"
	"sync"
)

var (
	cache sync.Map
)

func getRegexp(pattern string) *regexp.Regexp {
	if v, ok := cache.Load(pattern); ok {
		return v.(*regexp.Regexp)
	}

	reg := regexp.MustCompile(pattern)
	cache.Store(pattern, reg)
	return reg
}

func GetRegexpMatches(str, pattern string) []string {
	reg := getRegexp(pattern)
	matches := reg.FindAllStringSubmatch(str, -1)
	rets := make([]string, len(matches))
	for i, match := range matches {
		if len(match) > 1 {
			rets[i] = match[1]
		} else {
			rets[i] = match[0]
		}
	}
	return rets
}

func RegexpReplaceAll(str, pattern, newStr string) string {
	reg := getRegexp(pattern)
	return reg.ReplaceAllString(str, newStr)
}
