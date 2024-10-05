package parser

import (
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/aattwwss/ihf-referee-rules/pdf"
	"github.com/aattwwss/ihf-referee-rules/token"
	"golang.org/x/exp/slices"
)

type Question struct {
	ID             int
	Text           string
	Choices        []Choice
	Rule           Rule
	QuestionNumber int
	References     []Reference
}

type Choice struct {
	ID         int
	QuestionID int
	Option     string
	Text       string
	IsAnswer   bool
}

type Rule struct {
	ID        string
	Name      string
	SortOrder int
}

type Reference struct {
	ID         int
	QuestionID int
	Text       string
}

// soome rule question number uses comma, while others uses colon
func ruleQuestionNumSeparator(r rune) bool {
	return r == ':' || r == '.'
}

func ParseQuestion(tokens []token.Token, answerMap map[string]map[int]AnswersAndReferences) []Question {
	var allQuestions []Question
	groups := groupByQuestions(tokens)
	for _, group := range groups {
		q, err := toQuestion(len(allQuestions)+1, group, answerMap)
		if err != nil {
			slog.Error("convert to question error", slog.String("error", err.Error()))
			return nil
		}
		allQuestions = append(allQuestions, *q)
	}
	return allQuestions
}

// given the raw question string, split into the rule,
// question number and the question text
func splitQuestion(s string) (Rule, int, string) {
	bracketIndex := strings.IndexRune(s, ')')
	var ruleId string
	var qString string
	var text string

	text = s[bracketIndex+1:]
	if strings.HasPrefix(s, "SAR") {
		ruleId = "SAR"
		qString = s[3:4]
	} else {
		s = s[0:bracketIndex]
		arr := strings.FieldsFunc(s, ruleQuestionNumSeparator)
		ruleId = arr[0]
		qString = arr[1]
	}
	n, _ := strconv.Atoi(qString)
	return Rule{ID: ruleId}, n, strings.TrimSpace(text)
}

// given the raw choice string, split into the option and text
func splitChoice(s string) (string, string) {
	arr := strings.SplitN(s, ")", 2)
	return strings.TrimSpace(arr[0]), strings.TrimSpace(arr[1])
}

// filter away unneeded tokens, keeping only the questions and their choices.
// put each question in their own group
func groupByQuestions(tokens []token.Token) [][]token.Token {
	var groups [][]token.Token
	var group []token.Token
	isStart := true
	for _, t := range tokens {
		if t.Type == token.PAGE_NUMBER || t.Type == token.RULE_NUMBER {
			continue
		}

		if t.Type == token.QUESTION_START {
			if !isStart {
				groups = append(groups, mergeFreeText(group))
				group = []token.Token{}
			} else {
				isStart = false
			}
		}

		if !isStart {
			group = append(group, t)
		}
	}
	// add in the remaining group after going through all the tokens
	groups = append(groups, mergeFreeText(group))
	return groups
}

// merge the free text token with their main token to form a single value
func mergeFreeText(tokens []token.Token) []token.Token {
	var merged []token.Token
	for _, t := range tokens {
		if t.Type == token.QUESTION_START || t.Type == token.CHOICE_START {
			merged = append(merged, t)
		}

		if t.Type == token.FREE_TEXT {
			last := merged[len(merged)-1]
			last.Value = fmt.Sprintf("%s %s", last.Value, t.Value)
			merged[len(merged)-1] = last
		}
	}
	return merged
}

// given a token group of question and choices, construct the question object
func toQuestion(id int, tokens []token.Token, answerMap map[string]map[int]AnswersAndReferences) (*Question, error) {
	var q Question
	var choices []Choice
	for _, t := range tokens {
		if t.Type == token.QUESTION_START {
			rule, qNum, text := splitQuestion(t.Value)
			q.ID = id
			q.Rule = rule
			q.QuestionNumber = qNum
			q.Text = text
		} else if t.Type == token.CHOICE_START {
			option, choiceText := splitChoice(t.Value)
			c := Choice{
				QuestionID: id,
				Option:     option,
				Text:       choiceText,
				IsAnswer:   slices.Contains(answerMap[q.Rule.ID][q.QuestionNumber].Answers, option),
			}
			choices = append(choices, c)
		} else {
			return nil, fmt.Errorf("token not recognised: %s", t.Type)
		}
	}
	q.Choices = choices

	var references []Reference
	for _, r := range answerMap[q.Rule.ID][q.QuestionNumber].References {
		references = append(references, Reference{
			QuestionID: id,
			Text:       r,
		})
	}
	q.References = references
	return &q, nil
}

type AnswersAndReferences struct {
	Answers    []string
	References []string
}

// ParseAnswer returns the list of answers and references for a given rule and question number
func ParseAnswer(file io.Reader) map[string]map[int]AnswersAndReferences {
	ansMap := map[string]map[int]AnswersAndReferences{}
	s, _ := pdf.PdfToText(file)
	for _, s := range strings.Split(s, "\n") {
		s = strings.TrimSpace(s)
		if !hasAnswers(s) {
			continue
		}
		rule, questionNum, answers, references := splitAnswer(s)
		ruleMap, ok := ansMap[rule]
		if ok {
			ruleMap[questionNum] = AnswersAndReferences{
				Answers:    answers,
				References: references,
			}
		} else {
			ansMap[rule] = map[int]AnswersAndReferences{
				questionNum: {
					Answers:    answers,
					References: references,
				},
			}
		}
	}
	return ansMap
}

func hasAnswers(s string) bool {
	pattern := `\d\)`
	regex, _ := regexp.Compile(pattern)
	return regex.MatchString(s)
}

// Split the answer into the rule number, question number, the list of answers and the references
func splitAnswer(s string) (string, int, []string, []string) {
	// split into chunks of ruleNumber.QuestionNumber, answers, and references
	fields := regexp.MustCompile(` {2,}`).Split(s, -1)
	rule, questionNumber := getRuleQuestionNum(fields[0])
	correctAnswers := strings.Split(fields[1], ", ")
	references := strings.Split(fields[2], ", ")
	return rule, questionNumber, correctAnswers, references
}

// given the rule and question number, return the rule and the question number
// 18.7) -> "18", 7
// SAR1 -> "SAR", 1
func getRuleQuestionNum(s string) (string, int) {
	// time away the close bracket
	s = strings.TrimRight(s, ")")
	rule := ""
	questionNumberString := ""
	if strings.HasPrefix(s, "SAR") {
		// SAR does not have the period to separate the question number
		rule = "SAR"
		questionNumberString = strings.TrimLeft(s, "SAR")
	} else {
		arr := strings.FieldsFunc(s, ruleQuestionNumSeparator)
		rule = arr[0]
		questionNumberString = arr[1]
	}
	questionNumber, _ := strconv.Atoi(questionNumberString)
	return rule, questionNumber
}
