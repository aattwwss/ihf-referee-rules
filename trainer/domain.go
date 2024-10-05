package trainer

type QuestionEntity struct {
	ID             int
	Text           string
	RuleID         string
	QuestionNumber int
}

type ChoiceEntity struct {
	ID         int
	QuestionId int
	Option     string
	Text       string
	IsAnswer   bool
}

type RuleEntity struct {
	ID        string
	Name      string
	SortOrder int
}

type ReferenceEntity struct {
	ID         int
	QuestionId int
	Text       string
}

type FeedbackEntity struct {
	ID             int
	Name           string
	Email          string
	Topic          string
	Text           string
	IsAcknowledged bool
	IsCompleted    bool
}

type QuizConfigEntity struct {
	ID                 int
	Key                string
	NumQuestions       int
	DurationInMinutes  int
	HasNegativeMarking bool
	Seed               int
}

type QuizConfigRuleEntity struct {
	ID           int
	QuizConfigID int
	RuleID       string
}

type Question struct {
	ID                 int
	Text               string
	Rule               Rule
	QuestionNumber     int
	RuleQuestionNumber string
	Choices            []Choice
	References         []Reference
}

// ChoiceResult is used in the template to represent the result of a choice
// it should correspond to the class in the CSS
type ChoiceResult string

const (
	ChoiceResultCorrect ChoiceResult = "correct"
	ChoiceResultWrong   ChoiceResult = "wrong"
	ChoiceResultMissing ChoiceResult = "missing"
)

func (cr ChoiceResult) String() string {
	return string(cr)
}

func (cr ChoiceResult) CalcScore() int {
	switch cr {
	case ChoiceResultCorrect:
		return 1
	case ChoiceResultWrong:
		return -1
	default:
		return 0
	}
}

type Choice struct {
	ID         int
	Option     string
	Text       string
	IsAnswer   bool
	IsSelected bool
	Result     *ChoiceResult
}

type Rule struct {
	ID        string
	Name      string
	SortOrder int
}

type Reference struct {
	ID   int
	Text string
}

type Feedback struct {
	ID             int
	Name           string
	Email          string
	Topic          string
	Text           string
	IsAcknowledged bool
	IsCompleted    bool
}

type QuizConfig struct {
	ID                 int
	Key                string
	NumQuestions       int
	DurationInMinutes  int
	HasNegativeMarking bool
	Seed               int
	RuleIDs            []string
}

type QuestionResult struct {
	Question
	TotalScore int
	Score      int
}
