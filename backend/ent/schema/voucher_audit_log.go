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

// VoucherAuditLog records voucher order lifecycle events.
type VoucherAuditLog struct {
	ent.Schema
}

func (VoucherAuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "voucher_audit_logs"},
	}
}

func (VoucherAuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("order_id"),
		field.String("action").
			MaxLen(64),
		field.String("operator").
			MaxLen(128),
		field.JSON("metadata", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (VoucherAuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("created_at"),
	}
}
