package service

import (
	"context"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	// TicketStatusOpen: the user is waiting for staff.
	TicketStatusOpen = "open"
	// TicketStatusAnswered: staff (or AI) replied, waiting for the user.
	TicketStatusAnswered = "answered"
	TicketStatusClosed   = "closed"
)

const (
	TicketAuthorUser  = "user"
	TicketAuthorStaff = "staff"
	// TicketAuthorAI is reserved for automated replies.
	TicketAuthorAI = "ai"
)

const (
	ticketSubjectMaxLen = 200
	ticketContentMaxLen = 5000
)

var ticketCategories = map[string]struct{}{
	"account": {}, "billing": {}, "usage": {}, "other": {},
}

var (
	ErrTicketNotFound        = infraerrors.NotFound("TICKET_NOT_FOUND", "ticket not found")
	ErrTicketInvalidSubject  = infraerrors.BadRequest("TICKET_SUBJECT_INVALID", "ticket subject must be 1-200 characters")
	ErrTicketInvalidContent  = infraerrors.BadRequest("TICKET_CONTENT_INVALID", "ticket content must be 1-5000 characters")
	ErrTicketInvalidStatus   = infraerrors.BadRequest("TICKET_STATUS_INVALID", "ticket status is invalid")
	ErrTicketInvalidCategory = infraerrors.BadRequest("TICKET_CATEGORY_INVALID", "ticket category is invalid")
)

type Ticket struct {
	ID            int64
	UserID        int64
	UserEmail     string
	Subject       string
	Category      string
	Status        string
	LastMessageAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TicketMessage struct {
	ID         int64
	TicketID   int64
	AuthorType string
	AuthorID   *int64
	Content    string
	CreatedAt  time.Time
}

type TicketListFilters struct {
	// UserID limits the list to one user's tickets; nil lists everyone's.
	UserID *int64
	Status string
}

type TicketRepository interface {
	Create(ctx context.Context, ticket *Ticket, first *TicketMessage) (*Ticket, error)
	GetByID(ctx context.Context, id int64) (*Ticket, error)
	List(ctx context.Context, params pagination.PaginationParams, filters TicketListFilters) ([]Ticket, *pagination.PaginationResult, error)
	ListMessages(ctx context.Context, ticketID int64) ([]TicketMessage, error)
	// AddMessage stores the message and moves the ticket to status in one transaction.
	AddMessage(ctx context.Context, message *TicketMessage, status string) (*TicketMessage, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}
