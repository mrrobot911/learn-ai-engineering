package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
	"github.com/mrrobot911/learn-ai-engineering/internal/session"
)

// ContentHandler handles the public content endpoints.
type ContentHandler struct {
	service *service.ContentService
}

// NewContentHandler creates a content handler.
func NewContentHandler(contentService *service.ContentService) *ContentHandler {
	return &ContentHandler{service: contentService}
}

type topicSummaryResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int    `json:"position"`
	Status      string `json:"status"`
}

type topicDetailResponse struct {
	topicSummaryResponse
	Lessons []lessonResponse `json:"lessons"`
}

type lessonResponse struct {
	ID       string `json:"id"`
	TopicID  string `json:"topic_id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Position int    `json:"position"`
	Status   string `json:"status"`
}

type quizQuestionResponse struct {
	ID      string               `json:"id"`
	Text    string               `json:"text"`
	Options []quizOptionResponse `json:"options"`
}

type quizOptionResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type quizSubmissionRequest struct {
	Answers []quizSubmissionAnswerRequest `json:"answers"`
}

type quizSubmissionAnswerRequest struct {
	QuestionID string  `json:"question_id"`
	OptionID   *string `json:"option_id"`
}

type quizSubmissionResultResponse struct {
	Score   int                              `json:"score"`
	Total   int                              `json:"total"`
	Answers []quizSubmissionResultAnswerItem `json:"answers"`
}

type quizSubmissionResultAnswerItem struct {
	QuestionID        string `json:"question_id"`
	IsCorrect         bool   `json:"is_correct"`
	CorrectOptionText string `json:"correct_option_text"`
	Explanation       string `json:"explanation"`
}

type quizAttemptListItemResponse struct {
	ID        string    `json:"id"`
	TopicID   string    `json:"topic_id"`
	TopicSlug string    `json:"topic_slug"`
	Score     int       `json:"score"`
	Total     int       `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

type quizAttemptDetailResponse struct {
	quizAttemptListItemResponse
	Answers []quizAttemptAnswerResponse `json:"answers"`
}

type quizAttemptAnswerResponse struct {
	QuestionID         string `json:"question_id"`
	QuestionText       string `json:"question_text"`
	SelectedOptionText string `json:"selected_option_text"`
	CorrectOptionText  string `json:"correct_option_text"`
	IsCorrect          bool   `json:"is_correct"`
}

type draftLessonRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Position int    `json:"position"`
}

type draftLessonResponse struct {
	ID       string `json:"id"`
	TopicID  string `json:"topic_id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Position int    `json:"position"`
	Status   string `json:"status"`
}

type draftQuestionOptionInput struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

type draftQuestionRequest struct {
	Text    string                     `json:"text"`
	Options []draftQuestionOptionInput `json:"options"`
}

type draftQuestionOptionResponse struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

type draftQuestionResponse struct {
	ID      string                        `json:"id"`
	TopicID string                        `json:"topic_id"`
	Text    string                        `json:"text"`
	Options []draftQuestionOptionResponse `json:"options"`
	Status  string                        `json:"status"`
}

// ListTopics writes the published topics in position order.
func (h *ContentHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.service.ListTopics(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	response := make([]topicSummaryResponse, 0, len(topics))
	for _, topic := range topics {
		response = append(response, newTopicSummaryResponse(topic))
	}
	writeJSON(w, http.StatusOK, response)
}

// GetTopic writes one published topic with its published lessons in position
// order.
func (h *ContentHandler) GetTopic(w http.ResponseWriter, r *http.Request) {
	topic, lessons, err := h.service.GetTopic(r.Context(), r.PathValue("topicSlug"))
	if err != nil {
		WriteError(w, err)
		return
	}

	response := topicDetailResponse{
		topicSummaryResponse: newTopicSummaryResponse(topic),
		Lessons:              make([]lessonResponse, 0, len(lessons)),
	}
	for _, lesson := range lessons {
		response.Lessons = append(response.Lessons, newLessonResponse(lesson))
	}
	writeJSON(w, http.StatusOK, response)
}

// GetLesson writes one published lesson with its Markdown body.
func (h *ContentHandler) GetLesson(w http.ResponseWriter, r *http.Request) {
	lesson, err := h.service.GetLesson(r.Context(), r.PathValue("lessonId"))
	if err != nil {
		WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newLessonResponse(lesson))
}

// StartQuiz writes the questions of a quiz on the topic, with their options in
// position order and no field that reveals the answer. It works with and
// without a session.
func (h *ContentHandler) StartQuiz(w http.ResponseWriter, r *http.Request) {
	questions, err := h.service.StartQuiz(r.Context(), r.PathValue("topicSlug"))
	if err != nil {
		WriteError(w, err)
		return
	}

	response := make([]quizQuestionResponse, 0, len(questions))
	for _, question := range questions {
		options := make([]quizOptionResponse, 0, len(question.Options))
		for _, option := range question.Options {
			options = append(options, quizOptionResponse{ID: option.ID, Text: option.Text})
		}
		response = append(response, quizQuestionResponse{
			ID:      question.ID,
			Text:    question.Text,
			Options: options,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

// SubmitQuiz grades a quiz submission on the server. It works with or without
// a session; an anonymous submission returns the result and stores nothing.
func (h *ContentHandler) SubmitQuiz(w http.ResponseWriter, r *http.Request) {
	var request quizSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, &apperr.Error{Code: apperr.CodeValidation, Message: "The request body is invalid."})
		return
	}

	answers := make([]service.QuizAnswer, 0, len(request.Answers))
	for _, answer := range request.Answers {
		answers = append(answers, service.QuizAnswer{
			QuestionID: answer.QuestionID,
			OptionID:   answer.OptionID,
		})
	}

	userID := ""
	if auth, ok := session.From(r.Context()); ok {
		userID = auth.User.ID
	}

	result, err := h.service.SubmitQuiz(r.Context(), r.PathValue("topicSlug"), answers, userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	response := quizSubmissionResultResponse{
		Score:   result.Score,
		Total:   result.Total,
		Answers: make([]quizSubmissionResultAnswerItem, 0, len(result.Answers)),
	}
	for _, answer := range result.Answers {
		response.Answers = append(response.Answers, quizSubmissionResultAnswerItem{
			QuestionID:        answer.QuestionID,
			IsCorrect:         answer.IsCorrect,
			CorrectOptionText: answer.CorrectOptionText,
			Explanation:       answer.Explanation,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

// ListMyAttempts writes the stored quiz attempts of the logged-in user, newest
// first. It takes no user id from the request.
func (h *ContentHandler) ListMyAttempts(w http.ResponseWriter, r *http.Request) {
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{Code: apperr.CodeAuthentication, Message: "Authentication is required."})
		return
	}

	attempts, err := h.service.ListAttempts(r.Context(), auth.User.ID)
	if err != nil {
		WriteError(w, err)
		return
	}

	response := make([]quizAttemptListItemResponse, 0, len(attempts))
	for _, attempt := range attempts {
		response = append(response, newQuizAttemptListItemResponse(attempt))
	}
	writeJSON(w, http.StatusOK, response)
}

// GetMyAttempt writes one stored quiz attempt of the logged-in user. It takes no
// user id from the request.
func (h *ContentHandler) GetMyAttempt(w http.ResponseWriter, r *http.Request) {
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{Code: apperr.CodeAuthentication, Message: "Authentication is required."})
		return
	}

	attempt, err := h.service.GetAttempt(r.Context(), auth.User.ID, r.PathValue("attemptId"))
	if err != nil {
		WriteError(w, err)
		return
	}

	response := quizAttemptDetailResponse{
		quizAttemptListItemResponse: newQuizAttemptListItemResponse(attempt),
		Answers:                     make([]quizAttemptAnswerResponse, 0, len(attempt.Answers)),
	}
	for _, answer := range attempt.Answers {
		response.Answers = append(response.Answers, quizAttemptAnswerResponse{
			QuestionID:         answer.QuestionID,
			QuestionText:       answer.QuestionText,
			SelectedOptionText: answer.SelectedOptionText,
			CorrectOptionText:  answer.CorrectOptionText,
			IsCorrect:          answer.IsCorrect,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

// CreateDraftLesson creates a draft lesson for a topic.
func (h *ContentHandler) CreateDraftLesson(w http.ResponseWriter, r *http.Request) {
	var request draftLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, &apperr.Error{Code: apperr.CodeValidation, Message: "The request body is invalid."})
		return
	}

	lesson, err := h.service.CreateDraftLesson(r.Context(), r.PathValue("topicSlug"), service.DraftLessonInput{
		Title:    request.Title,
		Body:     request.Body,
		Position: request.Position,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, newDraftLessonResponse(lesson))
}

// UpdateDraftLesson updates an existing draft lesson.
func (h *ContentHandler) UpdateDraftLesson(w http.ResponseWriter, r *http.Request) {
	var request draftLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, &apperr.Error{Code: apperr.CodeValidation, Message: "The request body is invalid."})
		return
	}

	lesson, err := h.service.UpdateDraftLesson(r.Context(), r.PathValue("topicSlug"), r.PathValue("lessonId"), service.DraftLessonInput{
		Title:    request.Title,
		Body:     request.Body,
		Position: request.Position,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newDraftLessonResponse(lesson))
}

// CreateDraftQuestion creates a draft question for a topic.
func (h *ContentHandler) CreateDraftQuestion(w http.ResponseWriter, r *http.Request) {
	var request draftQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, &apperr.Error{Code: apperr.CodeValidation, Message: "The request body is invalid."})
		return
	}

	options := make([]service.DraftQuestionOptionInput, 0, len(request.Options))
	for _, opt := range request.Options {
		options = append(options, service.DraftQuestionOptionInput{
			Text:      opt.Text,
			IsCorrect: opt.IsCorrect,
		})
	}

	question, err := h.service.CreateDraftQuestion(r.Context(), r.PathValue("topicSlug"), service.DraftQuestionInput{
		Text:    request.Text,
		Options: options,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, newDraftQuestionResponse(question))
}

// UpdateDraftQuestion updates an existing draft question.
func (h *ContentHandler) UpdateDraftQuestion(w http.ResponseWriter, r *http.Request) {
	var request draftQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, &apperr.Error{Code: apperr.CodeValidation, Message: "The request body is invalid."})
		return
	}

	options := make([]service.DraftQuestionOptionInput, 0, len(request.Options))
	for _, opt := range request.Options {
		options = append(options, service.DraftQuestionOptionInput{
			Text:      opt.Text,
			IsCorrect: opt.IsCorrect,
		})
	}

	question, err := h.service.UpdateDraftQuestion(r.Context(), r.PathValue("topicSlug"), r.PathValue("questionId"), service.DraftQuestionInput{
		Text:    request.Text,
		Options: options,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newDraftQuestionResponse(question))
}

func newDraftLessonResponse(lesson repo.Lesson) draftLessonResponse {
	return draftLessonResponse{
		ID:       lesson.ID,
		TopicID:  lesson.TopicID,
		Title:    lesson.Title,
		Body:     lesson.Body,
		Position: lesson.Position,
		Status:   string(lesson.Status),
	}
}

func newDraftQuestionResponse(question repo.Question) draftQuestionResponse {
	options := make([]draftQuestionOptionResponse, 0, len(question.Options))
	for _, opt := range question.Options {
		options = append(options, draftQuestionOptionResponse{
			ID:        opt.ID,
			Text:      opt.Text,
			IsCorrect: opt.IsCorrect,
		})
	}
	return draftQuestionResponse{
		ID:      question.ID,
		TopicID: question.TopicID,
		Text:    question.Text,
		Options: options,
		Status:  string(question.Status),
	}
}

func newQuizAttemptListItemResponse(attempt repo.QuizAttempt) quizAttemptListItemResponse {
	return quizAttemptListItemResponse{
		ID:        attempt.ID,
		TopicID:   attempt.TopicID,
		TopicSlug: attempt.TopicSlug,
		Score:     attempt.Score,
		Total:     attempt.Total,
		CreatedAt: attempt.CreatedAt,
	}
}

func newTopicSummaryResponse(topic repo.Topic) topicSummaryResponse {
	return topicSummaryResponse{
		ID:          topic.ID,
		Slug:        topic.Slug,
		Title:       topic.Title,
		Description: topic.Description,
		Position:    topic.Position,
		Status:      string(topic.Status),
	}
}

func newLessonResponse(lesson repo.Lesson) lessonResponse {
	return lessonResponse{
		ID:       lesson.ID,
		TopicID:  lesson.TopicID,
		Title:    lesson.Title,
		Body:     lesson.Body,
		Position: lesson.Position,
		Status:   string(lesson.Status),
	}
}
