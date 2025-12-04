package customUtils

import "regexp"

func GetRegexpMatches(str, pattern string) []string {
	reg := regexp.MustCompile(pattern)
	matches := reg.FindAllStringSubmatch(str, -1)
	rets := make([]string, len(matches))
	for i, match := range matches {
		rets[i] = match[1]
	}
	return rets
}

func RegexpReplaceAll(str, pattern, newStr string) string {
	reg := regexp.MustCompile(pattern)
	return reg.ReplaceAllString(str, newStr)
}
