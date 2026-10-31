package handler

import (
	"net/http"

	"github.com/mrrobot911/learn-ai-engineering/internal/repo"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
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
