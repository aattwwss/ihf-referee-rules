package trainer

import (
	"context"
	"slices"
	"strings"

	"github.com/sethvargo/go-diceware/diceware"
	"golang.org/x/exp/rand"
)

type Repository interface {
	GetAllQuestions(ctx context.Context) ([]Question, error)
	GetAllRules(ctx context.Context) ([]Rule, error)
	GetQuestionByID(ctx context.Context, id int) (*Question, error)
	GetRandomQuestion(ctx context.Context, rules []string) (*Question, error)
	GetChoicesByQuestionID(ctx context.Context, questionID int) ([]Choice, error)
	InsertFeedback(ctx context.Context, feedback Feedback) error
	InsertQuizConfig(ctx context.Context, config QuizConfig, questions []Question) error
	GetQuizConfigByKey(ctx context.Context, key string) (*QuizConfig, error)
	GetQuestionsByQuizConfigKey(ctx context.Context, key string) ([]Question, error)
	FindReferencesByQuestionIds(ctx context.Context, questionIds ...int) (map[int][]Reference, error)
}

type QuestionService struct {
	repository Repository
}

func NewService(repository Repository) *QuestionService {
	return &QuestionService{repository: repository}
}

func (s *QuestionService) GetAllQuestions(ctx context.Context) ([]Question, error) {
	return s.repository.GetAllQuestions(ctx)
}

func (s *QuestionService) GetAllRules(ctx context.Context) ([]Rule, error) {
	return s.repository.GetAllRules(ctx)
}

func (s *QuestionService) GetQuestionByID(ctx context.Context, id int) (*Question, error) {
	return s.repository.GetQuestionByID(ctx, id)
}

func (s *QuestionService) GetRandomQuestion(ctx context.Context, rules []string) (*Question, error) {
	return s.repository.GetRandomQuestion(ctx, rules)
}

func (s *QuestionService) GetChoicesByQuestionID(ctx context.Context, questionID int) ([]Choice, error) {
	return s.repository.GetChoicesByQuestionID(ctx, questionID)
}

func (s *QuestionService) SubmitFeedback(ctx context.Context, feedback Feedback) error {
	return s.repository.InsertFeedback(ctx, feedback)
}

func (s *QuestionService) SubmitQuizConfig(ctx context.Context, quizConfig QuizConfig) (string, error) {
	key, err := s.generateQuizConfigKey(ctx)
	if err != nil {
		return "", err
	}
	quizConfig.Key = key
	allQuestions, err := s.repository.GetAllQuestions(ctx)
	if err != nil {
		return "", err
	}

	var questions []Question
	for _, q := range allQuestions {
		if slices.Contains(quizConfig.RuleIDs, q.Rule.ID) {
			questions = append(questions, q)
		}
	}

	rand.Seed(uint64(quizConfig.Seed))
	rand.Shuffle(len(questions), func(i, j int) { questions[i], questions[j] = questions[j], questions[i] })
	if len(questions) > quizConfig.NumQuestions {
		questions = questions[:quizConfig.NumQuestions]
	}
	slices.SortFunc(questions, func(a, b Question) int {
		if a.Rule.SortOrder == b.Rule.SortOrder {
			return a.QuestionNumber - b.QuestionNumber
		}
		return a.Rule.SortOrder - b.Rule.SortOrder
	})
	err = s.repository.InsertQuizConfig(ctx, quizConfig, questions)
	if err != nil {
		return "", err
	}
	return quizConfig.Key, nil
}

// GetQuestionsByQuizConfigKey returns a list of questions for the quiz based on the quiz config key
func (s *QuestionService) GetQuestionsByQuizConfigKey(ctx context.Context, key string) ([]Question, error) {
	return s.repository.GetQuestionsByQuizConfigKey(ctx, key)
}

func (s *QuestionService) EvaluateQuizAnswer(ctx context.Context, quizConfigKey string, choiceIDs []int) ([]QuestionResult, error) {
	quizConfig, err := s.repository.GetQuizConfigByKey(ctx, quizConfigKey)
	if err != nil {
		return nil, err
	}
	questions, err := s.GetQuestionsByQuizConfigKey(ctx, quizConfigKey)
	if err != nil {
		return nil, err
	}
	var questionIDs []int
	for _, question := range questions {
		questionIDs = append(questionIDs, question.ID)
	}
	var questionResults []QuestionResult
	for _, q := range questions {
		totalScore := 0
		score := 0
		for i, c := range q.Choices {
			isSelected := slices.Contains(choiceIDs, c.ID)
			if isSelected && c.IsAnswer {
				q.Choices[i].Result = refOf(ChoiceResultCorrect)
				totalScore += 1
			} else if isSelected && !c.IsAnswer {
				q.Choices[i].Result = refOf(ChoiceResultWrong)
			} else if !isSelected && c.IsAnswer {
				q.Choices[i].Result = refOf(ChoiceResultMissing)
				totalScore += 1
			}
			if q.Choices[i].Result != nil {
				score += q.Choices[i].Result.CalcScore()
			}
		}
		if score < 0 && !quizConfig.HasNegativeMarking {
			score = 0
		}
		questionResults = append(questionResults, QuestionResult{
			Question:   q,
			TotalScore: totalScore,
			Score:      score,
		})
	}
	return questionResults, nil
}

// refOf returns a pointer to the given value
func refOf[E any](e E) *E {
	return &e
}

func (s *QuestionService) generateQuizConfigKey(ctx context.Context) (string, error) {
	list, err := diceware.Generate(3)
	if err != nil {
		return "", err
	}
	key := strings.Join(list, "-")
	configInDB, err := s.repository.GetQuizConfigByKey(ctx, key)
	if err != nil {
		return "", err
	}
	if configInDB != nil {
		return s.generateQuizConfigKey(ctx)
	}
	return key, nil
}
