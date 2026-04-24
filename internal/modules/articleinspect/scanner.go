package articleinspect

import (
	"context"
	"regexp"
	"strings"
	"time"
)

type CandidateArticle struct {
	ID            uint64
	OrgID         uint64
	Title         string
	ShortTitle    string
	RichTitle     string
	Keyword       string
	Desc          string
	Body          string
	State         int8
	PublishAtTime *time.Time
}

type KeywordRule struct {
	ID            uint64   `json:"id"`
	Name          string   `json:"name"`
	Category      string   `json:"category_name"`
	MatchType     string   `json:"match_type"`
	RiskLevel     string   `json:"risk_level"`
	SuggestAction string   `json:"suggest_action"`
	Scopes        []string `json:"scopes"`
}

type Hit struct {
	KeywordID     uint64
	KeywordText   string
	Category      string
	FieldName     string
	MatchType     string
	RiskLevel     string
	SuggestAction string
	MatchedText   string
	Snippet       string
	PositionStart int
	PositionEnd   int
}

type Scanner interface {
	ScanArticle(ctx context.Context, article CandidateArticle, rules []KeywordRule) ([]Hit, error)
}

type KeywordScanner struct {
	SnippetContext int
	MaxRegexLength int
}

func NewKeywordScanner() *KeywordScanner {
	return &KeywordScanner{
		SnippetContext: 24,
		MaxRegexLength: 128,
	}
}

func (s *KeywordScanner) ScanArticle(ctx context.Context, article CandidateArticle, rules []KeywordRule) ([]Hit, error) {
	_ = ctx
	if s == nil {
		s = NewKeywordScanner()
	}

	hits := make([]Hit, 0)
	for _, rule := range rules {
		for _, scope := range rule.Scopes {
			value := articleFieldValue(article, scope)
			if strings.TrimSpace(value) == "" {
				continue
			}
			hits = append(hits, s.scanField(rule, scope, value)...)
		}
	}
	return hits, nil
}

func (s *KeywordScanner) scanField(rule KeywordRule, fieldName, value string) []Hit {
	matchType := strings.ToLower(strings.TrimSpace(rule.MatchType))
	switch matchType {
	case MatchTypeContains:
		return s.containsHits(rule, fieldName, value)
	case MatchTypeExact:
		return s.exactHits(rule, fieldName, value)
	case MatchTypeRegex:
		return s.regexHits(rule, fieldName, value)
	default:
		return nil
	}
}

func (s *KeywordScanner) containsHits(rule KeywordRule, fieldName, value string) []Hit {
	needle := strings.TrimSpace(rule.Name)
	if needle == "" {
		return nil
	}
	lowerValue := strings.ToLower(value)
	lowerNeedle := strings.ToLower(needle)
	matches := make([]Hit, 0)
	searchFrom := 0
	for {
		index := strings.Index(lowerValue[searchFrom:], lowerNeedle)
		if index < 0 {
			break
		}
		start := searchFrom + index
		end := start + len(needle)
		matches = append(matches, buildHit(rule, fieldName, value, value[start:end], start, end, s.SnippetContext))
		searchFrom = end
	}
	return matches
}

func (s *KeywordScanner) exactHits(rule KeywordRule, fieldName, value string) []Hit {
	needle := strings.TrimSpace(rule.Name)
	field := strings.TrimSpace(value)
	if needle == "" || field == "" || !strings.EqualFold(field, needle) {
		return nil
	}
	return []Hit{buildHit(rule, fieldName, value, field, 0, len(field), s.SnippetContext)}
}

func (s *KeywordScanner) regexHits(rule KeywordRule, fieldName, value string) []Hit {
	pattern := strings.TrimSpace(rule.Name)
	if pattern == "" || !isSafeRegexPattern(pattern, s.MaxRegexLength) {
		return nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	locs := re.FindAllStringIndex(value, -1)
	matches := make([]Hit, 0, len(locs))
	for _, loc := range locs {
		if len(loc) != 2 {
			continue
		}
		matches = append(matches, buildHit(rule, fieldName, value, value[loc[0]:loc[1]], loc[0], loc[1], s.SnippetContext))
	}
	return matches
}

func articleFieldValue(article CandidateArticle, scope string) string {
	switch scope {
	case KeywordScopeTitle:
		return article.Title
	case KeywordScopeShortTitle:
		return article.ShortTitle
	case KeywordScopeRichTitle:
		return article.RichTitle
	case KeywordScopeKeyword:
		return article.Keyword
	case KeywordScopeDesc:
		return article.Desc
	case KeywordScopeBody:
		return article.Body
	default:
		return ""
	}
}

func buildHit(rule KeywordRule, fieldName, fullText, matched string, start, end, contextWindow int) Hit {
	return Hit{
		KeywordID:     rule.ID,
		KeywordText:   rule.Name,
		Category:      rule.Category,
		FieldName:     fieldName,
		MatchType:     rule.MatchType,
		RiskLevel:     rule.RiskLevel,
		SuggestAction: rule.SuggestAction,
		MatchedText:   matched,
		Snippet:       buildSnippet(fullText, start, end, contextWindow),
		PositionStart: start,
		PositionEnd:   end,
	}
}

func buildSnippet(text string, start, end, contextWindow int) string {
	if text == "" {
		return ""
	}
	if contextWindow <= 0 {
		contextWindow = 24
	}
	snippetStart := start - contextWindow
	if snippetStart < 0 {
		snippetStart = 0
	}
	snippetEnd := end + contextWindow
	if snippetEnd > len(text) {
		snippetEnd = len(text)
	}

	prefix := ""
	suffix := ""
	if snippetStart > 0 {
		prefix = "..."
	}
	if snippetEnd < len(text) {
		suffix = "..."
	}
	return prefix + text[snippetStart:snippetEnd] + suffix
}

func isSafeRegexPattern(pattern string, maxLength int) bool {
	if maxLength <= 0 {
		maxLength = 128
	}
	if len(pattern) > maxLength {
		return false
	}
	return !strings.Contains(pattern, "(?")
}
