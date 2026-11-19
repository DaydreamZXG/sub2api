package repository

import (
	"context"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/ticket"
	"github.com/Wei-Shaw/sub2api/ent/ticketmessage"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type ticketRepository struct {
	client *dbent.Client
}

func NewTicketRepository(client *dbent.Client) service.TicketRepository {
	return &ticketRepository{client: client}
}

// withTx runs fn in a transaction, joining the caller's when one is already open.
func (r *ticketRepository) withTx(ctx context.Context, fn func(client *dbent.Client) error) error {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return fn(tx.Client())
	}
	tx, err := r.client.Tx(ctx)
	if errors.Is(err, dbent.ErrTxStarted) {
		return fn(r.client)
	}
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx.Client()); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ticketRepository) Create(ctx context.Context, t *service.Ticket, first *service.TicketMessage) (*service.Ticket, error) {
	var created *dbent.Ticket
	err := r.withTx(ctx, func(client *dbent.Client) error {
		var err error
		created, err = client.Ticket.Create().
			SetUserID(t.UserID).
			SetSubject(t.Subject).
			SetCategory(t.Category).
			SetStatus(t.Status).
			Save(ctx)
		if err != nil {
			return err
		}
		return client.TicketMessage.Create().
			SetTicketID(created.ID).
			SetAuthorType(first.AuthorType).
			SetNillableAuthorID(first.AuthorID).
			SetContent(first.Content).
			Exec(ctx)
	})
	if err != nil {
		return nil, err
	}
	return ticketEntityToService(created), nil
}

func (r *ticketRepository) GetByID(ctx context.Context, id int64) (*service.Ticket, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.Ticket.Query().Where(ticket.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTicketNotFound, nil)
	}
	out := ticketEntityToService(m)
	email, err := client.User.Query().Where(user.IDEQ(m.UserID)).Select(user.FieldEmail).String(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return nil, err
	}
	out.UserEmail = email
	return out, nil
}

func (r *ticketRepository) List(ctx context.Context, params pagination.PaginationParams, filters service.TicketListFilters) ([]service.Ticket, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.Ticket.Query()
	if filters.UserID != nil {
		q = q.Where(ticket.UserIDEQ(*filters.UserID))
	}
	if filters.Status != "" {
		q = q.Where(ticket.StatusEQ(filters.Status))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	items, err := q.
		Order(dbent.Desc(ticket.FieldLastMessageAt), dbent.Desc(ticket.FieldID)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	userIDs := make([]int64, 0, len(items))
	for _, item := range items {
		userIDs = append(userIDs, item.UserID)
	}
	users, err := client.User.Query().Where(user.IDIn(userIDs...)).Select(user.FieldID, user.FieldEmail).All(ctx)
	if err != nil {
		return nil, nil, err
	}
	emails := make(map[int64]string, len(users))
	for _, u := range users {
		emails[u.ID] = u.Email
	}

	out := make([]service.Ticket, 0, len(items))
	for _, item := range items {
		t := ticketEntityToService(item)
		t.UserEmail = emails[item.UserID]
		out = append(out, *t)
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (r *ticketRepository) ListMessages(ctx context.Context, ticketID int64) ([]service.TicketMessage, error) {
	client := clientFromContext(ctx, r.client)
	items, err := client.TicketMessage.Query().
		Where(ticketmessage.TicketIDEQ(ticketID)).
		Order(dbent.Asc(ticketmessage.FieldCreatedAt), dbent.Asc(ticketmessage.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.TicketMessage, 0, len(items))
	for _, item := range items {
		out = append(out, *ticketMessageEntityToService(item))
	}
	return out, nil
}

func (r *ticketRepository) AddMessage(ctx context.Context, message *service.TicketMessage, status string) (*service.TicketMessage, error) {
	var created *dbent.TicketMessage
	err := r.withTx(ctx, func(client *dbent.Client) error {
		var err error
		created, err = client.TicketMessage.Create().
			SetTicketID(message.TicketID).
			SetAuthorType(message.AuthorType).
			SetNillableAuthorID(message.AuthorID).
			SetContent(message.Content).
			Save(ctx)
		if err != nil {
			return err
		}
		updated, err := client.Ticket.Update().
			Where(ticket.IDEQ(message.TicketID)).
			SetStatus(status).
			SetLastMessageAt(time.Now()).
			Save(ctx)
		if err != nil {
			return err
		}
		if updated == 0 {
			return service.ErrTicketNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ticketMessageEntityToService(created), nil
}

func (r *ticketRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	client := clientFromContext(ctx, r.client)
	updated, err := client.Ticket.Update().Where(ticket.IDEQ(id)).SetStatus(status).Save(ctx)
	if err != nil {
		return err
	}
	if updated == 0 {
		return service.ErrTicketNotFound
	}
	return nil
}

func ticketEntityToService(m *dbent.Ticket) *service.Ticket {
	return &service.Ticket{
		ID:            m.ID,
		UserID:        m.UserID,
		Subject:       m.Subject,
		Category:      m.Category,
		Status:        m.Status,
		LastMessageAt: m.LastMessageAt,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func ticketMessageEntityToService(m *dbent.TicketMessage) *service.TicketMessage {
	return &service.TicketMessage{
		ID:         m.ID,
		TicketID:   m.TicketID,
		AuthorType: m.AuthorType,
		AuthorID:   m.AuthorID,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt,
	}
}
