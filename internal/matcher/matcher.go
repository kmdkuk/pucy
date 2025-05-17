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

func (m Matches) MatchCount() int {
	count := 0
	for _, match := range m {
		count += match.last - match.first
	}
	return count
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
	lineRunes := []rune(lowerLine)

	// Split keyword by space and ignore empty keywords
	keywords := strings.Fields(strings.ToLower(keyword))

	for _, kw := range keywords {
		if kw == "" {
			continue
		}
		keywordRunes := []rune(kw)
		for i := 0; i <= len(lineRunes)-len(keywordRunes); i++ {
			match := true
			for j := 0; j < len(keywordRunes); j++ {
				if lineRunes[i+j] != keywordRunes[j] {
					match = false
					break
				}
			}
			if match {
				matches = append(matches, Match{first: i, last: i + len(keywordRunes)})
			}
		}
	}

	return matches
}
