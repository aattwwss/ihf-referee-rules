package trainer

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type QuestionRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *QuestionRepository {
	return &QuestionRepository{
		db: db,
	}
}

func (r *QuestionRepository) GetAllQuestions(ctx context.Context) ([]Question, error) {
	query := fmt.Sprintf(`
		SELECT q.id, q.text, q.rule_id, q.question_number 
		FROM question q join rule r on q.rule_id = r.id 
		ORDER BY r.sort_order, q.question_number 
	`)
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	questionEntities, err := pgx.CollectRows(rows, pgx.RowToStructByPos[QuestionEntity])
	if err != nil {
		return nil, err
	}
	questionIds := make([]int, 0, len(questionEntities))
	for _, questionEntity := range questionEntities {
		questionIds = append(questionIds, questionEntity.ID)
	}
	choiceMap, err := r.FindChoicesByQuestionIds(ctx, questionIds...)
	referencesMap, err := r.FindReferencesByQuestionIds(ctx, questionIds...)
	ruleIDs, err := r.GetAllDistinctRuleIDs(ctx)
	rulesMap, err := r.FindRuleByIDs(ctx, ruleIDs...)
	var questions []Question
	for _, questionEntity := range questionEntities {
		separator := "."
		if questionEntity.RuleID == "SAR" {
			separator = ""
		}
		ruleQuestionNumber := fmt.Sprintf("%s%s%d", questionEntity.RuleID, separator, questionEntity.QuestionNumber)
		question := Question{
			ID:                 questionEntity.ID,
			Text:               questionEntity.Text,
			Rule:               rulesMap[questionEntity.RuleID],
			QuestionNumber:     questionEntity.QuestionNumber,
			RuleQuestionNumber: ruleQuestionNumber,
			Choices:            choiceMap[questionEntity.ID],
			References:         referencesMap[questionEntity.ID],
		}
		questions = append(questions, question)
	}
	return questions, nil
}

func (r *QuestionRepository) GetQuestionByID(ctx context.Context, id int) (*Question, error) {
	query := fmt.Sprintf("SELECT id, text, rule_id, question_number FROM question WHERE id = $1")
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	questionEntity, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[QuestionEntity])
	if err != nil {
		return nil, err
	}
	query = fmt.Sprintf("SELECT * FROM choice WHERE question_id = $1 order by option")
	rows, err = r.db.Query(ctx, query, questionEntity.ID)
	if err != nil {
		return nil, err
	}

	choiceEntities, err := pgx.CollectRows(rows, pgx.RowToStructByPos[ChoiceEntity])
	var choices []Choice
	for _, choiceEntity := range choiceEntities {
		choices = append(choices, Choice{
			ID:         choiceEntity.ID,
			Option:     choiceEntity.Option,
			Text:       choiceEntity.Text,
			IsAnswer:   choiceEntity.IsAnswer,
			IsSelected: false,
		})
	}
	separator := "."
	if questionEntity.RuleID == "SAR" {
		separator = ""
	}
	ruleQuestionNumber := fmt.Sprintf("%s%s%d", questionEntity.RuleID, separator, questionEntity.QuestionNumber)
	rule, err := r.FindRuleByID(ctx, questionEntity.RuleID)
	if err != nil {
		return nil, err
	}
	return &Question{
		ID:                 questionEntity.ID,
		Text:               questionEntity.Text,
		Rule:               *rule,
		QuestionNumber:     questionEntity.QuestionNumber,
		RuleQuestionNumber: ruleQuestionNumber,
		Choices:            choices,
	}, nil
}

func (r *QuestionRepository) GetRandomQuestion(ctx context.Context, rules []string) (*Question, error) {
	if len(rules) == 0 {
		allRules, err := r.GetAllDistinctRuleIDs(ctx)
		if err != nil {
			return nil, err
		}
		rules = allRules
	}
	query := fmt.Sprintf("SELECT id, text, rule_id, question_number FROM question WHERE rule_id = ANY($1) ORDER BY RANDOM() LIMIT 1")
	rows, err := r.db.Query(ctx, query, rules)
	if err != nil {
		return nil, err
	}
	questionEntity, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[QuestionEntity])
	if err != nil {
		return nil, err
	}
	query = fmt.Sprintf("SELECT * FROM choice WHERE question_id = $1 order by option")
	rows, err = r.db.Query(ctx, query, questionEntity.ID)
	if err != nil {
		return nil, err
	}

	choiceEntities, err := pgx.CollectRows(rows, pgx.RowToStructByPos[ChoiceEntity])
	var choices []Choice
	for _, choiceEntity := range choiceEntities {
		choices = append(choices, Choice{
			ID:         choiceEntity.ID,
			Option:     choiceEntity.Option,
			Text:       choiceEntity.Text,
			IsAnswer:   choiceEntity.IsAnswer,
			IsSelected: false,
		})
	}
	separator := "."
	if questionEntity.RuleID == "SAR" {
		separator = ""
	}
	ruleQuestionNumber := fmt.Sprintf("%s%s%d", questionEntity.RuleID, separator, questionEntity.QuestionNumber)
	rule, err := r.FindRuleByID(ctx, questionEntity.RuleID)
	if err != nil {
		return nil, err
	}
	return &Question{
		ID:                 questionEntity.ID,
		Text:               questionEntity.Text,
		Rule:               *rule,
		QuestionNumber:     questionEntity.QuestionNumber,
		RuleQuestionNumber: ruleQuestionNumber,
		Choices:            choices,
	}, nil
}

func (r *QuestionRepository) GetChoicesByQuestionID(ctx context.Context, questionID int) ([]Choice, error) {
	query := fmt.Sprintf("SELECT * FROM choice WHERE question_id = $1 order by option")
	rows, err := r.db.Query(ctx, query, questionID)
	if err != nil {
		return nil, err
	}
	choiceEntities, err := pgx.CollectRows(rows, pgx.RowToStructByPos[ChoiceEntity])
	if err != nil {
		return nil, err
	}
	var choices []Choice
	for _, choiceEntity := range choiceEntities {
		choices = append(choices, Choice{
			ID:         choiceEntity.ID,
			Option:     choiceEntity.Option,
			Text:       choiceEntity.Text,
			IsAnswer:   choiceEntity.IsAnswer,
			IsSelected: false,
		})
	}
	return choices, nil
}

func (r *QuestionRepository) FindRuleByID(ctx context.Context, ruleID string) (*Rule, error) {
	rulesMap, err := r.FindRuleByIDs(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if rule, ok := rulesMap[ruleID]; ok {
		return &rule, nil
	}
	return nil, nil
}

// FindRuleByIDs finds rules by rule ids and returns a map of rule id to rule
func (r *QuestionRepository) FindRuleByIDs(ctx context.Context, ruleIDs ...string) (map[string]Rule, error) {
	query := fmt.Sprintf("SELECT * FROM rule WHERE id = ANY($1)")
	rows, err := r.db.Query(ctx, query, ruleIDs)
	if err != nil {
		return nil, err
	}
	ruleEntity, err := pgx.CollectRows(rows, pgx.RowToStructByPos[RuleEntity])
	if err != nil {
		return nil, err
	}
	var rulesMap = make(map[string]Rule)
	for _, ruleEntity := range ruleEntity {
		rulesMap[ruleEntity.ID] = Rule{
			ID:        ruleEntity.ID,
			Name:      ruleEntity.Name,
			SortOrder: ruleEntity.SortOrder,
		}
	}
	return rulesMap, nil
}

func (r *QuestionRepository) GetAllDistinctRuleIDs(ctx context.Context) ([]string, error) {

	query := fmt.Sprintf("SELECT id FROM rule")
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	// scan int string array
	var rules []string
	for rows.Next() {

		var rule string
		err = rows.Scan(&rule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *QuestionRepository) GetAllRules(ctx context.Context) ([]Rule, error) {
	query := fmt.Sprintf("SELECT * FROM rule order by sort_order")
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	ruleEntity, err := pgx.CollectRows(rows, pgx.RowToStructByPos[RuleEntity])
	if err != nil {
		return nil, err
	}

	var rules []Rule
	for _, ruleEntity := range ruleEntity {
		rules = append(rules, Rule{
			ID:        ruleEntity.ID,
			Name:      ruleEntity.Name,
			SortOrder: ruleEntity.SortOrder,
		})
	}
	return rules, nil
}

// ListQuestions returns a list of questions
// supports pagination using the rule sort order and question number of the last question to offset
func (r *QuestionRepository) ListQuestions(ctx context.Context, questionIDs []int, ruleIDs []string, search string, lastRuleSortOrder int, lastQuestionNumber int, limit int) ([]Question, error) {
	if len(ruleIDs) == 0 {
		allRules, err := r.GetAllDistinctRuleIDs(ctx)
		if err != nil {
			return nil, err
		}
		ruleIDs = allRules
	}
	query := fmt.Sprintf(`
		SELECT q.id, q.text, q.rule_id, q.question_number 
		FROM question q join rule r on q.rule_id = r.id 
		WHERE r.id = ANY($1) 
			AND ($2 = '' OR tsv @@ websearch_to_tsquery($2))
			AND r.sort_order >= $3
			AND q.question_number > $4
			AND q.id = ANY($5)
		ORDER BY r.sort_order, q.question_number 
		LIMIT $6
	`)
	rows, err := r.db.Query(ctx, query, ruleIDs, strings.TrimSpace(search), lastRuleSortOrder, lastQuestionNumber, questionIDs, limit)
	if err != nil {
		return nil, err
	}
	questionEntities, err := pgx.CollectRows(rows, pgx.RowToStructByPos[QuestionEntity])
	if err != nil {
		return nil, err
	}
	questionIds := make([]int, 0, len(questionEntities))
	for _, questionEntity := range questionEntities {
		questionIds = append(questionIds, questionEntity.ID)
	}
	choiceMap, err := r.FindChoicesByQuestionIds(ctx, questionIds...)
	rulesMap, err := r.FindRuleByIDs(ctx, ruleIDs...)
	referencesMap, err := r.FindReferencesByQuestionIds(ctx, questionIds...)
	var questions []Question
	for _, questionEntity := range questionEntities {
		separator := "."
		if questionEntity.RuleID == "SAR" {
			separator = ""
		}
		ruleQuestionNumber := fmt.Sprintf("%s%s%d", questionEntity.RuleID, separator, questionEntity.QuestionNumber)
		question := Question{
			ID:                 questionEntity.ID,
			Text:               questionEntity.Text,
			Rule:               rulesMap[questionEntity.RuleID],
			QuestionNumber:     questionEntity.QuestionNumber,
			RuleQuestionNumber: ruleQuestionNumber,
			Choices:            choiceMap[questionEntity.ID],
			References:         referencesMap[questionEntity.ID],
		}
		questions = append(questions, question)
	}
	return questions, nil
}

// FindChoicesByQuestionIds finds choices by question ids and returns a map of question id to choices
func (r *QuestionRepository) FindChoicesByQuestionIds(ctx context.Context, questionIds ...int) (map[int][]Choice, error) {
	query := fmt.Sprintf("SELECT * FROM choice WHERE question_id = ANY($1) order by option")
	rows, err := r.db.Query(ctx, query, questionIds)
	if err != nil {
		return nil, err
	}
	choiceEntities, err := pgx.CollectRows(rows, pgx.RowToStructByPos[ChoiceEntity])
	if err != nil {
		return nil, err
	}
	var choiceMap = make(map[int][]Choice)
	for _, choiceEntity := range choiceEntities {
		choices, ok := choiceMap[choiceEntity.QuestionId]
		if !ok {
			choices = []Choice{}
		}

		choices = append(choices, Choice{
			ID:         choiceEntity.ID,
			Option:     choiceEntity.Option,
			Text:       choiceEntity.Text,
			IsAnswer:   choiceEntity.IsAnswer,
			IsSelected: false,
		})
		choiceMap[choiceEntity.QuestionId] = choices
	}
	return choiceMap, nil
}

// FindReferencesByQuestionIds finds references by question ids and returns a map of question id to references
func (r *QuestionRepository) FindReferencesByQuestionIds(ctx context.Context, questionIds ...int) (map[int][]Reference, error) {
	query := fmt.Sprintf("SELECT * FROM reference WHERE question_id = ANY($1) order by id")
	rows, err := r.db.Query(ctx, query, questionIds)
	if err != nil {
		return nil, err
	}
	referenceEntities, err := pgx.CollectRows(rows, pgx.RowToStructByPos[ReferenceEntity])
	if err != nil {
		return nil, err
	}
	var referenceMap = make(map[int][]Reference)
	for _, referenceEntity := range referenceEntities {
		references, ok := referenceMap[referenceEntity.QuestionId]
		if !ok {
			references = []Reference{}
		}

		references = append(references, Reference{
			ID:   referenceEntity.ID,
			Text: referenceEntity.Text,
		})
		referenceMap[referenceEntity.QuestionId] = references
	}
	return referenceMap, nil
}

func (r *QuestionRepository) InsertFeedback(ctx context.Context, feedback Feedback) error {
	feedbackEntity := FeedbackEntity{
		Name:           feedback.Name,
		Email:          feedback.Email,
		Topic:          feedback.Topic,
		Text:           feedback.Text,
		IsAcknowledged: feedback.IsAcknowledged,
		IsCompleted:    feedback.IsCompleted,
	}
	query := fmt.Sprintf("INSERT INTO feedback (email, Name, topic, text, is_acknowledged, is_completed) VALUES ($1, $2, $3, $4,$5, $6)")
	_, err := r.db.Exec(ctx, query, feedbackEntity.Email, feedbackEntity.Name, feedbackEntity.Topic, feedbackEntity.Text, feedbackEntity.IsAcknowledged, feedbackEntity.IsCompleted)
	if err != nil {
		return err
	}
	return nil
}

func (r *QuestionRepository) InsertQuizConfig(ctx context.Context, quizConfig QuizConfig, questions []Question) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	entity := QuizConfigEntity{
		Key:                quizConfig.Key,
		NumQuestions:       quizConfig.NumQuestions,
		DurationInMinutes:  quizConfig.DurationInMinutes,
		HasNegativeMarking: quizConfig.HasNegativeMarking,
		Seed:               quizConfig.Seed,
	}
	query := fmt.Sprintf("INSERT INTO quiz_config (key, num_questions, duration_in_minutes, has_negative_marking, seed) VALUES ($1, $2, $3, $4, $5) RETURNING id")
	err = tx.QueryRow(ctx, query, entity.Key, entity.NumQuestions, entity.DurationInMinutes, entity.HasNegativeMarking, entity.Seed).Scan(&entity.ID)
	if err != nil {
		return err
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"quiz_config_rules"},
		[]string{"quiz_config_id", "rule_id"},
		pgx.CopyFromSlice(len(quizConfig.RuleIDs), func(i int) ([]interface{}, error) {
			return []interface{}{entity.ID, quizConfig.RuleIDs[i]}, nil
		}),
	)
	if err != nil {
		return err
	}
	var questionIDs []int
	for _, question := range questions {
		questionIDs = append(questionIDs, question.ID)
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"quiz_config_questions"},
		[]string{"quiz_config_id", "question_id"},
		pgx.CopyFromSlice(len(questionIDs), func(i int) ([]interface{}, error) {
			return []interface{}{entity.ID, questionIDs[i]}, nil
		}),
	)
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *QuestionRepository) GetQuizConfigByKey(ctx context.Context, key string) (*QuizConfig, error) {
	query := fmt.Sprintf("SELECT * FROM quiz_config WHERE key = $1")
	rows, err := r.db.Query(ctx, query, key)
	if err != nil {
		return nil, err
	}
	quizConfigEntity, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[QuizConfigEntity])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	query = fmt.Sprintf("SELECT rule_id FROM quiz_config_rules WHERE quiz_config_id = $1")
	rows, err = r.db.Query(ctx, query, quizConfigEntity.ID)
	if err != nil {
		return nil, err
	}
	var ruleIDs []string
	for rows.Next() {
		var ruleID string
		err = rows.Scan(&ruleID)
		if err != nil {
			return nil, err
		}
		ruleIDs = append(ruleIDs, ruleID)
	}

	return &QuizConfig{
		ID:                 quizConfigEntity.ID,
		Key:                quizConfigEntity.Key,
		NumQuestions:       quizConfigEntity.NumQuestions,
		DurationInMinutes:  quizConfigEntity.DurationInMinutes,
		HasNegativeMarking: quizConfigEntity.HasNegativeMarking,
		Seed:               quizConfigEntity.Seed,
		RuleIDs:            ruleIDs,
	}, nil
}

func (r *QuestionRepository) GetQuestionsByQuizConfigKey(ctx context.Context, key string) ([]Question, error) {
	query := fmt.Sprintf("SELECT * FROM quiz_config WHERE key = $1")
	rows, err := r.db.Query(ctx, query, key)
	if err != nil {
		return nil, err
	}
	quizConfigEntity, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[QuizConfigEntity])
	if err != nil {
		return nil, err
	}

	query = fmt.Sprintf("SELECT question_id FROM quiz_config_questions c join question q on c.question_id = q.id WHERE quiz_config_id = $1")
	rows, err = r.db.Query(ctx, query, quizConfigEntity.ID)
	if err != nil {
		return nil, err
	}
	var questionIDs []int
	for rows.Next() {
		var questionID int
		err = rows.Scan(&questionID)
		if err != nil {
			return nil, err
		}
		questionIDs = append(questionIDs, questionID)
	}
	return r.ListQuestions(ctx, questionIDs, nil, "", 0, 0, len(questionIDs))
}
