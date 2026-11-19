package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type Ticket struct {
	ID            int64     `json:"id"`
	Subject       string    `json:"subject"`
	Category      string    `json:"category"`
	Status        string    `json:"status"`
	LastMessageAt time.Time `json:"last_message_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// AdminTicket adds who opened the ticket; only admin endpoints return it.
type AdminTicket struct {
	Ticket
	UserID    int64  `json:"user_id"`
	UserEmail string `json:"user_email"`
}

type TicketMessage struct {
	ID         int64     `json:"id"`
	AuthorType string    `json:"author_type"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type TicketDetail struct {
	Ticket
	Messages []TicketMessage `json:"messages"`
}

type AdminTicketDetail struct {
	AdminTicket
	Messages []TicketMessage `json:"messages"`
}

func TicketFromService(t *service.Ticket) Ticket {
	return Ticket{
		ID:            t.ID,
		Subject:       t.Subject,
		Category:      t.Category,
		Status:        t.Status,
		LastMessageAt: t.LastMessageAt,
		CreatedAt:     t.CreatedAt,
	}
}

func AdminTicketFromService(t *service.Ticket) AdminTicket {
	return AdminTicket{Ticket: TicketFromService(t), UserID: t.UserID, UserEmail: t.UserEmail}
}

func TicketMessageFromService(m *service.TicketMessage) TicketMessage {
	return TicketMessage{ID: m.ID, AuthorType: m.AuthorType, Content: m.Content, CreatedAt: m.CreatedAt}
}

func TicketMessagesFromService(messages []service.TicketMessage) []TicketMessage {
	out := make([]TicketMessage, 0, len(messages))
	for i := range messages {
		out = append(out, TicketMessageFromService(&messages[i]))
	}
	return out
}
