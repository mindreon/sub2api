package schema

import (
	"fmt"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MediaQuotaHold 多模态任务额度预扣台账（reserve / settle / release）。
type MediaQuotaHold struct {
	ent.Schema
}

func (MediaQuotaHold) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "media_quota_holds"},
	}
}

func (MediaQuotaHold) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (MediaQuotaHold) Fields() []ent.Field {
	return []ent.Field{
		field.String("hold_id").
			MaxLen(64).
			Unique(),
		field.String("task_id").
			MaxLen(64).
			Comment("关联 media_generation_tasks.task_id"),
		field.Int64("user_id"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.String("currency").
			MaxLen(8).
			Default("USD"),
		field.String("status").
			MaxLen(16).
			Default("held").
			Validate(validateMediaHoldStatus),
	}
}

func (MediaQuotaHold) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id").Unique(),
		index.Fields("user_id", "status"),
	}
}

func validateMediaHoldStatus(status string) error {
	switch status {
	case "held", "settled", "released":
		return nil
	default:
		return fmt.Errorf("invalid media hold status: %s", status)
	}
}
