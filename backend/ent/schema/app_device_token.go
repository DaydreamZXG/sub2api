package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AppDeviceToken holds the schema definition for the AppDeviceToken entity.
//
// 桌面助手在某台设备上的专属凭证：网关据此判断一个请求是否由已登录的 app 转发。
// 只保存哈希，原文只在签发时返回一次。
type AppDeviceToken struct {
	ent.Schema
}

func (AppDeviceToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "app_device_tokens"},
	}
}

func (AppDeviceToken) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("device_id").MaxLen(128).NotEmpty(),
		field.String("platform").MaxLen(16).Default("unknown"),
		field.String("client_version").MaxLen(32).Default(""),
		field.String("token_hash").MaxLen(64).NotEmpty().Unique(),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("revoked_at").Optional().Nillable(),
	}
}

func (AppDeviceToken) Indexes() []ent.Index {
	return []ent.Index{
		// 一台设备同时只有一个有效凭证，并发重连也不例外。
		index.Fields("user_id", "device_id").
			Unique().
			Annotations(entsql.IndexWhere("revoked_at IS NULL")),
	}
}
