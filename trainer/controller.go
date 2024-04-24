package trainer

import (
	"context"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type Service interface {
	GetRandomQuestion(ctx context.Context, rules []string) (*Question, error)
	GetChoicesByQuestionID(ctx context.Context, questionID int) ([]Choice, error)
	ListQuestions(ctx context.Context, rules []string, search string, lastRuleSortOrder int, lastQuestionNumber int, limit int) ([]Question, error)
}

type Controller struct {
	service Service
	html    fs.FS
}

func NewController(service Service, html fs.FS) *Controller {
	return &Controller{
		service: service,
		html:    html,
	}
}

func (c *Controller) Home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(c.html, "base.tmpl", "home.tmpl")
	if err != nil {
		log.Printf("Error parsing template: %s", err)
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing template: %s", err)
	}
}

func (c *Controller) QuestionList(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(c.html, "questionList.tmpl")
	if err != nil {
		log.Printf("Error parsing template: %s", err)
	}
	search := queryParamString(r, "search", "goal-area")
	lastRuleSortOrder := queryParamInt(r, "lastRuleSortOrder", 0)
	lastQuestionNumber := queryParamInt(r, "lastQuestionNumber", 0)
	questions, err := c.service.ListQuestions(r.Context(), nil, search, lastRuleSortOrder, lastQuestionNumber, 10)
	if err != nil {
		log.Printf("Error getting questions: %s", err)
	}
	err = tmpl.Execute(w, questions)
	if err != nil {
		log.Printf("Error executing template: %s", err)
	}
}

func (c *Controller) RandomQuestion(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.ParseFS(c.html, "base.tmpl", "randomQuestion.tmpl")
	if err != nil {
		log.Printf("Error parsing template: %s", err)
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing template: %s", err)
	}
}

func (c *Controller) NewQuestion(w http.ResponseWriter, r *http.Request) {
	rules, err := getQueryStrings(r, "rules")
	if err != nil {
		log.Printf("Error getting query strings: %s", err)
	}
	question, err := c.service.GetRandomQuestion(r.Context(), rules)
	if err != nil {
		log.Printf("Error getting random question: %s", err)
	}
	tmpl, err := template.ParseFS(c.html, "question.tmpl")
	if err != nil {
		log.Printf("Error parsing template: %s", err)
	}
	err = tmpl.Execute(w, question)
	if err != nil {
		log.Printf("Error executing template: %s", err)
	}
}

func (c *Controller) Result(w http.ResponseWriter, r *http.Request) {
	questionID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/submit/"))
	if err != nil {
		log.Printf("Error parsing question ID: %s", err)
	}
	err = r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %s", err)
	}
	selected := r.Form["choices"]
	choices, err := c.service.GetChoicesByQuestionID(r.Context(), questionID)
	for i, choice := range choices {
		if slices.Contains(selected, choice.Option) {
			choices[i].IsSelected = true
		}
	}
	if err != nil {
		log.Printf("Error getting choices: %s", err)
	}

	tmpl, err := template.ParseFS(c.html, "result.tmpl")
	if err != nil {
		log.Printf("Error parsing template: %s", err)
	}
	err = tmpl.Execute(w, choices)
	if err != nil {
		log.Printf("Error executing template: %s", err)
	}
}

func (c *Controller) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// getQueryStrings parses a list of query strings from the request.
func getQueryStrings(r *http.Request, query string) ([]string, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	var ss []string
	for _, s := range r.Form[query] {
		ss = append(ss, strings.Split(s, ",")...)
	}
	return ss, nil
}

func queryParamString(r *http.Request, query string, defaultValue string) string {
	if s := r.URL.Query().Get(query); s != "" {
		return s
	}
	return defaultValue
}

func queryParamInt(r *http.Request, query string, defaultValue int) int {
	s := r.URL.Query().Get(query)
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Print("Error parsing int: ", err)
		return defaultValue
	}
	return i
}
