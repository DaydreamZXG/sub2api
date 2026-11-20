package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type TicketService struct {
	repo TicketRepository
}

func NewTicketService(repo TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

type CreateTicketInput struct {
	Subject  string
	Category string
	Content  string
}

func normalizeTicketContent(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" || utf8.RuneCountInString(content) > ticketContentMaxLen {
		return "", ErrTicketInvalidContent
	}
	return content, nil
}

// Create opens a ticket for userID with its first message.
func (s *TicketService) Create(ctx context.Context, userID int64, input CreateTicketInput) (*Ticket, error) {
	subject := strings.TrimSpace(input.Subject)
	if subject == "" || utf8.RuneCountInString(subject) > ticketSubjectMaxLen {
		return nil, ErrTicketInvalidSubject
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "other"
	}
	if _, ok := ticketCategories[category]; !ok {
		return nil, ErrTicketInvalidCategory
	}
	content, err := normalizeTicketContent(input.Content)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx,
		&Ticket{UserID: userID, Subject: subject, Category: category, Status: TicketStatusOpen},
		&TicketMessage{AuthorType: TicketAuthorUser, AuthorID: &userID, Content: content},
	)
}

func (s *TicketService) ListForUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Ticket, *pagination.PaginationResult, error) {
	return s.repo.List(ctx, params, TicketListFilters{UserID: &userID})
}

// getOwned hides other users' tickets behind the same error as a missing one.
func (s *TicketService) getOwned(ctx context.Context, userID, ticketID int64) (*Ticket, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.UserID != userID {
		return nil, ErrTicketNotFound
	}
	return ticket, nil
}

func (s *TicketService) GetForUser(ctx context.Context, userID, ticketID int64) (*Ticket, []TicketMessage, error) {
	ticket, err := s.getOwned(ctx, userID, ticketID)
	if err != nil {
		return nil, nil, err
	}
	messages, err := s.repo.ListMessages(ctx, ticketID)
	if err != nil {
		return nil, nil, err
	}
	return ticket, messages, nil
}

// ReplyAsUser adds the user's message and puts the ticket back in the staff queue,
// reopening it when it was closed.
func (s *TicketService) ReplyAsUser(ctx context.Context, userID, ticketID int64, content string) (*TicketMessage, error) {
	content, err := normalizeTicketContent(content)
	if err != nil {
		return nil, err
	}
	if _, err := s.getOwned(ctx, userID, ticketID); err != nil {
		return nil, err
	}
	return s.repo.AddMessage(ctx,
		&TicketMessage{TicketID: ticketID, AuthorType: TicketAuthorUser, AuthorID: &userID, Content: content},
		TicketStatusOpen,
	)
}

func (s *TicketService) CloseByUser(ctx context.Context, userID, ticketID int64) error {
	if _, err := s.getOwned(ctx, userID, ticketID); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, ticketID, TicketStatusClosed)
}

func (s *TicketService) List(ctx context.Context, params pagination.PaginationParams, status string) ([]Ticket, *pagination.PaginationResult, error) {
	if status != "" && !isTicketStatus(status) {
		return nil, nil, ErrTicketInvalidStatus
	}
	return s.repo.List(ctx, params, TicketListFilters{Status: status})
}

func (s *TicketService) Get(ctx context.Context, ticketID int64) (*Ticket, []TicketMessage, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, nil, err
	}
	messages, err := s.repo.ListMessages(ctx, ticketID)
	if err != nil {
		return nil, nil, err
	}
	return ticket, messages, nil
}

// ReplyAsStaff adds a staff message and hands the ticket back to the user.
func (s *TicketService) ReplyAsStaff(ctx context.Context, staffID, ticketID int64, content string) (*TicketMessage, error) {
	content, err := normalizeTicketContent(content)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetByID(ctx, ticketID); err != nil {
		return nil, err
	}
	return s.repo.AddMessage(ctx,
		&TicketMessage{TicketID: ticketID, AuthorType: TicketAuthorStaff, AuthorID: &staffID, Content: content},
		TicketStatusAnswered,
	)
}

func (s *TicketService) SetStatus(ctx context.Context, ticketID int64, status string) error {
	if !isTicketStatus(status) {
		return ErrTicketInvalidStatus
	}
	if _, err := s.repo.GetByID(ctx, ticketID); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, ticketID, status)
}

func isTicketStatus(status string) bool {
	return status == TicketStatusOpen || status == TicketStatusAnswered || status == TicketStatusClosed
}
