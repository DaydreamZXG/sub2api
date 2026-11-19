package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ticketRepoStub struct {
	tickets  map[int64]*Ticket
	messages []TicketMessage
}

func newTicketRepoStub() *ticketRepoStub {
	return &ticketRepoStub{tickets: map[int64]*Ticket{}}
}

func (r *ticketRepoStub) Create(_ context.Context, t *Ticket, first *TicketMessage) (*Ticket, error) {
	created := *t
	created.ID = int64(len(r.tickets) + 1)
	r.tickets[created.ID] = &created
	message := *first
	message.TicketID = created.ID
	r.messages = append(r.messages, message)
	return &created, nil
}

func (r *ticketRepoStub) GetByID(_ context.Context, id int64) (*Ticket, error) {
	t, ok := r.tickets[id]
	if !ok {
		return nil, ErrTicketNotFound
	}
	out := *t
	return &out, nil
}

func (r *ticketRepoStub) List(_ context.Context, _ pagination.PaginationParams, filters TicketListFilters) ([]Ticket, *pagination.PaginationResult, error) {
	var out []Ticket
	for _, t := range r.tickets {
		if filters.UserID != nil && t.UserID != *filters.UserID {
			continue
		}
		if filters.Status != "" && t.Status != filters.Status {
			continue
		}
		out = append(out, *t)
	}
	return out, &pagination.PaginationResult{Total: int64(len(out))}, nil
}

func (r *ticketRepoStub) ListMessages(_ context.Context, ticketID int64) ([]TicketMessage, error) {
	var out []TicketMessage
	for _, m := range r.messages {
		if m.TicketID == ticketID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *ticketRepoStub) AddMessage(_ context.Context, message *TicketMessage, status string) (*TicketMessage, error) {
	r.messages = append(r.messages, *message)
	r.tickets[message.TicketID].Status = status
	return message, nil
}

func (r *ticketRepoStub) UpdateStatus(_ context.Context, id int64, status string) error {
	r.tickets[id].Status = status
	return nil
}

func TestTicketServiceCreateValidatesInput(t *testing.T) {
	tests := []struct {
		name  string
		input CreateTicketInput
		want  error
	}{
		{"blank subject", CreateTicketInput{Subject: "  ", Content: "help"}, ErrTicketInvalidSubject},
		{"subject too long", CreateTicketInput{Subject: strings.Repeat("字", 201), Content: "help"}, ErrTicketInvalidSubject},
		{"blank content", CreateTicketInput{Subject: "s", Content: " \n "}, ErrTicketInvalidContent},
		{"content too long", CreateTicketInput{Subject: "s", Content: strings.Repeat("字", 5001)}, ErrTicketInvalidContent},
		{"unknown category", CreateTicketInput{Subject: "s", Category: "vip", Content: "help"}, ErrTicketInvalidCategory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTicketRepoStub()
			_, err := NewTicketService(repo).Create(context.Background(), 7, tt.input)
			require.ErrorIs(t, err, tt.want)
			assert.Empty(t, repo.tickets)
		})
	}
}

func TestTicketServiceCreateStoresTrimmedTicketWithFirstMessage(t *testing.T) {
	repo := newTicketRepoStub()

	// 200 multi-byte characters are within the limit: it counts characters, not bytes.
	subject := strings.Repeat("字", 200)
	ticket, err := NewTicketService(repo).Create(context.Background(), 7,
		CreateTicketInput{Subject: "  " + subject + " ", Content: " 连不上 \n"})
	require.NoError(t, err)

	assert.Equal(t, subject, ticket.Subject)
	assert.Equal(t, "other", ticket.Category)
	assert.Equal(t, TicketStatusOpen, ticket.Status)
	require.Len(t, repo.messages, 1)
	assert.Equal(t, "连不上", repo.messages[0].Content)
	assert.Equal(t, TicketAuthorUser, repo.messages[0].AuthorType)
	require.NotNil(t, repo.messages[0].AuthorID)
	assert.Equal(t, int64(7), *repo.messages[0].AuthorID)
}

func TestTicketServiceHidesOtherUsersTickets(t *testing.T) {
	repo := newTicketRepoStub()
	svc := NewTicketService(repo)
	ctx := context.Background()
	ticket, err := svc.Create(ctx, 7, CreateTicketInput{Subject: "s", Content: "c"})
	require.NoError(t, err)

	_, _, err = svc.GetForUser(ctx, 8, ticket.ID)
	require.ErrorIs(t, err, ErrTicketNotFound)
	_, err = svc.ReplyAsUser(ctx, 8, ticket.ID, "let me in")
	require.ErrorIs(t, err, ErrTicketNotFound)
	require.ErrorIs(t, svc.CloseByUser(ctx, 8, ticket.ID), ErrTicketNotFound)

	assert.Len(t, repo.messages, 1)
	assert.Equal(t, TicketStatusOpen, repo.tickets[ticket.ID].Status)

	items, _, err := svc.ListForUser(ctx, 8, pagination.PaginationParams{})
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestTicketServiceStatusFollowsWhoRepliedLast(t *testing.T) {
	repo := newTicketRepoStub()
	svc := NewTicketService(repo)
	ctx := context.Background()
	ticket, err := svc.Create(ctx, 7, CreateTicketInput{Subject: "s", Content: "c"})
	require.NoError(t, err)

	reply, err := svc.ReplyAsStaff(ctx, 1, ticket.ID, "try again")
	require.NoError(t, err)
	assert.Equal(t, TicketAuthorStaff, reply.AuthorType)
	assert.Equal(t, TicketStatusAnswered, repo.tickets[ticket.ID].Status)

	require.NoError(t, svc.CloseByUser(ctx, 7, ticket.ID))
	assert.Equal(t, TicketStatusClosed, repo.tickets[ticket.ID].Status)

	// A user reply on a closed ticket puts it back in the staff queue.
	_, err = svc.ReplyAsUser(ctx, 7, ticket.ID, "still broken")
	require.NoError(t, err)
	assert.Equal(t, TicketStatusOpen, repo.tickets[ticket.ID].Status)

	_, messages, err := svc.GetForUser(ctx, 7, ticket.ID)
	require.NoError(t, err)
	assert.Len(t, messages, 3)
}

func TestTicketServiceRejectsUnknownStatus(t *testing.T) {
	repo := newTicketRepoStub()
	svc := NewTicketService(repo)
	ctx := context.Background()
	ticket, err := svc.Create(ctx, 7, CreateTicketInput{Subject: "s", Content: "c"})
	require.NoError(t, err)

	require.ErrorIs(t, svc.SetStatus(ctx, ticket.ID, "archived"), ErrTicketInvalidStatus)
	_, _, err = svc.List(ctx, pagination.PaginationParams{}, "archived")
	require.ErrorIs(t, err, ErrTicketInvalidStatus)
	require.ErrorIs(t, svc.SetStatus(ctx, 999, TicketStatusClosed), ErrTicketNotFound)

	require.NoError(t, svc.SetStatus(ctx, ticket.ID, TicketStatusClosed))
	assert.Equal(t, TicketStatusClosed, repo.tickets[ticket.ID].Status)
}
