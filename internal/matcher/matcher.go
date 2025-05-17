package matcher

import "strings"

type Match struct {
	first int
	last  int
}

type Matches []Match

func (m Matches) IsMatch(index int) bool {
	for _, match := range m {
		if index >= match.first && index < match.last {
			return true
		}
	}
	return false
}

type Matcher interface {
	Match(string, string) Matches
}

type DefaultMatcher struct {
}

func NewMatcher() Matcher {
	return &DefaultMatcher{}
}
func (m *DefaultMatcher) Match(line string, keyword string) Matches {
	if keyword == "" {
		return nil
	}

	var matches Matches
	lowerLine := strings.ToLower(line)
	lowerKeyword := strings.ToLower(keyword)
	lineRunes := []rune(lowerLine)
	keywordRunes := []rune(lowerKeyword)

	for i := 0; i <= len(lineRunes)-len(keywordRunes); i++ {
		if string(lineRunes[i:i+len(keywordRunes)]) == string(keywordRunes) {
			matches = append(matches, Match{first: i, last: i + len(keywordRunes)})
		}
	}

	return matches
}
