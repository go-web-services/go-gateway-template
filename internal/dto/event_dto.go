package dto

import "time"

// EventSendRequestDTO is the JSON body from the browser for POST /api/v1/events/send.
// The gateway fills project_id, message_id (when omitted), ip, user_agent, and user_id when the session is authenticated.
type EventSendRequestDTO struct {
	DistinctID string         `json:"distinct_id" binding:"required" example:"anon_6ba7b810-9dad-11d1-80b4-00c04fd430c8"`
	SessionID  *string        `json:"session_id,omitempty" example:"sess_01j8xyz"`
	Name       string         `json:"name" binding:"required,max=255" example:"page_viewed"`
	Payload    map[string]any `json:"payload" binding:"required"`
	OccurredAt time.Time      `json:"occurred_at" binding:"required" example:"2026-04-12T10:30:00.123Z"`
	// MessageID is optional; when set it must be a UUID (idempotency key). Omit entirely or use a real UUID — not Swagger’s placeholder "string".
	MessageID string `json:"message_id,omitempty" binding:"omitempty"`
}

// EventSendOutputDTO is returned after a successful event ingest (no event body).
type EventSendOutputDTO struct {
	Message string `json:"message" example:"success"`
}
