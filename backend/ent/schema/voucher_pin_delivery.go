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

// VoucherPinDelivery stores PIN codes delivered for a voucher order.
type VoucherPinDelivery struct {
	ent.Schema
}

func (VoucherPinDelivery) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "voucher_pin_deliveries"},
	}
}

func (VoucherPinDelivery) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("order_id"),
		field.String("pin_code_enc").
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("serial").
			MaxLen(128).
			Default(""),
		field.Float("denomination").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("delivered_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (VoucherPinDelivery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
	}
}
