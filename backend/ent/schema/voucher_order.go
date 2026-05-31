package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// VoucherOrder is a user-facing PIN purchase order (bank transfer flow).
type VoucherOrder struct {
	ent.Schema
}

func (VoucherOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "voucher_orders"},
	}
}

func (VoucherOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").
			MaxLen(64).
			Unique(),
		field.Int64("user_id"),
		field.String("user_email").
			MaxLen(255),
		field.String("user_name").
			MaxLen(100),

		field.String("status").
			MaxLen(32).
			Default("pending_payment"),

		field.Int64("product_id"),
		field.Int64("kv_product_id").
			Optional().
			Nillable(),
		field.String("product_name").
			MaxLen(128),
		field.Float("denomination").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Int("quantity").
			Default(1),
		field.Float("unit_price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Float("subtotal").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Float("fee_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.Float("total_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.String("currency").
			MaxLen(3).
			Default("MYR"),

		field.String("payment_ref").
			Optional().
			Nillable().
			MaxLen(128),
		field.String("payment_proof_path").
			Optional().
			Nillable().
			MaxLen(512),
		field.Int("bank_account_id").
			Optional().
			Nillable(),

		field.String("kv_retrieve_reference").
			MaxLen(128),
		field.String("idempotency_key").
			Optional().
			Nillable().
			MaxLen(128),
		field.String("reject_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("fulfill_error").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		field.Time("expires_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("verified_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("fulfilled_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("completed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		field.String("client_ip").
			MaxLen(50).
			Default(""),

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

func (VoucherOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("voucher_orders").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (VoucherOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("created_at"),
		index.Fields("idempotency_key").
			Unique().
			Annotations(entsql.IndexWhere("idempotency_key IS NOT NULL AND idempotency_key <> ''")),
	}
}
