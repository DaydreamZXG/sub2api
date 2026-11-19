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

// Ticket holds the schema definition for the Ticket entity.
//
// 用户提交的工单：一个标题加一串对话（见 TicketMessage）。
type Ticket struct {
	ent.Schema
}

func (Ticket) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "tickets"},
	}
}

func (Ticket) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("subject").
			MaxLen(200).
			NotEmpty(),
		field.String("category").
			MaxLen(32).
			Default("other"),
		field.String("status").
			MaxLen(16).
			Default("open").
			Comment("open: 等待客服回复; answered: 已回复，等待用户; closed: 已关闭"),
		field.Time("last_message_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (Ticket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "last_message_at"),
		index.Fields("status", "last_message_at"),
	}
}
