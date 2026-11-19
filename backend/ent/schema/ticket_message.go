package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TicketMessage holds the schema definition for the TicketMessage entity.
//
// 工单里的一条消息。author_type 区分用户、客服和 AI，方便以后接入自动回复。
type TicketMessage struct {
	ent.Schema
}

func (TicketMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ticket_messages"},
	}
}

func (TicketMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("ticket_id"),
		field.String("author_type").
			MaxLen(16).
			Comment("user | staff | ai"),
		field.Int64("author_id").
			Optional().
			Nillable().
			Comment("发送者的用户 ID；AI 回复为空"),
		field.String("content").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			NotEmpty(),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (TicketMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ticket_id", "created_at"),
	}
}
