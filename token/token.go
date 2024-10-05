package token

import (
	"fmt"
	"regexp"
	"strings"
)

type Type int

const (
	PAGE_NUMBER Type = iota
	RULE_NUMBER
	QUESTION_START
	CHOICE_START
	FREE_TEXT
	IGNORE
)

func (tt Type) String() string {
	return []string{"PAGE_NUMBER", "RULE_NUMBER", "QUESTION_START", "CHOICE_START", "FREE_TEXT", "IGNORE"}[tt]

}

type Token struct {
	Type  Type
	Value string
}

type Pattern struct {
	Type  Type
	Match func(s string) bool
}

type Tokenizer struct {
	Matchers []Pattern
}

func NewTokenizer() Tokenizer {
	// if the line contains only the word Rule and a number
	ruleNumberMatcher := func(s string) bool {
		s = strings.TrimSpace(s)
		regex := regexp.MustCompile(`^Rule \d+$`)
		return regex.MatchString(s)
	}
	ruleNumberTokenPattern := Pattern{
		Type:  RULE_NUMBER,
		Match: ruleNumberMatcher,
	}

	// line contains the number only
	pageNumberMatcher := func(s string) bool {
		s = strings.TrimSpace(s)
		regex := regexp.MustCompile(`^\d+$`)
		return regex.MatchString(s)
	}
	pageNumberTokenPattern := Pattern{
		Type:  PAGE_NUMBER,
		Match: pageNumberMatcher,
	}

	// line starts with the rule number and question number follow by a close bracket
	// also account for the Substitution Area Regulator that starts with "SAR"
	// e.g. 12.34) some other text
	// e.g. 12:34) some other text
	// e.g. SAR1) some other text
	questionStartMatcher := func(s string) bool {
		s = strings.TrimSpace(s)
		regex := regexp.MustCompile(`^(\d+[.:]\d+\)|SAR\d\)) .+$`)
		return regex.MatchString(s)
	}
	questionStartTokenPattern := Pattern{
		Type:  QUESTION_START,
		Match: questionStartMatcher,
	}

	// line starts with the choice character follow by a close bracket
	// e.g. a) some other text
	choiceStartMatcher := func(s string) bool {
		s = strings.TrimSpace(s)
		regex := regexp.MustCompile(`^.\) .+$`)
		return regex.MatchString(s)
	}
	choiceStartTokenPattern := Pattern{
		Type:  CHOICE_START,
		Match: choiceStartMatcher,
	}

	// lines to ignore
	ignoreMatcher := func(s string) bool {
		return "Substitution Area Regulation" == strings.TrimSpace(s)
	}
	ignoreTokenPattern := Pattern{
		Type:  IGNORE,
		Match: ignoreMatcher,
	}

	// free text is anything else
	freeTextMatcher := func(s string) bool {
		return true
	}
	freeTextTokenPattern := Pattern{
		Type:  FREE_TEXT,
		Match: freeTextMatcher,
	}

	// return the tokenizer where the order of the token pattern matters!
	return Tokenizer{
		Matchers: []Pattern{
			ruleNumberTokenPattern,
			pageNumberTokenPattern,
			questionStartTokenPattern,
			choiceStartTokenPattern,
			ignoreTokenPattern,
			freeTextTokenPattern,
		},
	}
}

func (t Tokenizer) Tokenize(s string) (*Token, error) {
	s = strings.TrimSpace(s)
	for _, matcher := range t.Matchers {
		if matcher.Match(s) {
			return &Token{matcher.Type, s}, nil
		}
	}
	return nil, fmt.Errorf("no token found for %s", s)
}
