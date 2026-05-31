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

// VoucherB2BOrder is a platform wholesale replenishment order to KVoucher.
type VoucherB2BOrder struct {
	ent.Schema
}

func (VoucherB2BOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "voucher_b2b_orders"},
	}
}

func (VoucherB2BOrder) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("kv_order_id").
			Unique(),
		field.String("order_no").
			MaxLen(64).
			Unique(),
		field.String("status").
			MaxLen(32).
			Default("pending_payment"),
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
		field.JSON("items_json", []map[string]any{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("payment_info_json", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
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
		field.String("merchant_notes").
			Optional().
			Nillable().
			MaxLen(500),
		field.String("idempotency_key").
			Optional().
			Nillable().
			MaxLen(128),
		field.String("reject_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("kv_last_request_id").
			Optional().
			Nillable().
			MaxLen(64),
		field.String("created_by").
			MaxLen(128).
			Default("admin"),
		field.Time("kv_last_synced_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("verified_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("pins_loaded_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("completed_at").
			Optional().
			Nillable().
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

func (VoucherB2BOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("created_at"),
		index.Fields("idempotency_key").
			Unique().
			Annotations(entsql.IndexWhere("idempotency_key IS NOT NULL AND idempotency_key <> ''")),
	}
}
