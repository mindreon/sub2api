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

// VoucherProduct is a sellable PIN denomination synced from KVoucher.
type VoucherProduct struct {
	ent.Schema
}

func (VoucherProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "voucher_products"},
	}
}

func (VoucherProduct) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("kv_product_id").
			Optional().
			Nillable(),
		field.String("name").
			MaxLen(128),
		field.Float("denomination").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Float("wholesale_price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.Float("retail_price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.String("currency").
			MaxLen(3).
			Default("MYR"),
		field.Int("stock_available").
			Default(0),
		field.Bool("is_active").
			Default(true),
		field.Int("sort_order").
			Default(0),
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

func (VoucherProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kv_product_id").
			Unique().
			Annotations(entsql.IndexWhere("kv_product_id IS NOT NULL")),
		index.Fields("denomination"),
		index.Fields("is_active"),
	}
}
