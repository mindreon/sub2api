package schema

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MediaGenerationTask 多模态异步生成任务（视频/音频/图片）。
type MediaGenerationTask struct {
	ent.Schema
}

func (MediaGenerationTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "media_generation_tasks"},
	}
}

func (MediaGenerationTask) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (MediaGenerationTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("task_id").
			MaxLen(64).
			Unique().
			Comment("对外任务 ID"),
		field.String("upstream_task_id").
			Optional().
			Nillable().
			MaxLen(128).
			Comment("上游任务 ID"),
		field.Int64("user_id"),
		field.Int64("api_key_id"),
		field.Int64("account_id").
			Optional().
			Nillable(),
		field.Int64("group_id").
			Optional().
			Nillable(),
		field.Int64("subscription_id").
			Optional().
			Nillable().
			Comment("提交时快照的用户订阅 ID，用于异步结算保持订阅计费语义"),
		field.String("model").
			MaxLen(100),
		field.String("media_type").
			MaxLen(16).
			Comment("video/audio/image"),
		field.String("status").
			MaxLen(20).
			Default("pending").
			Validate(validateMediaTaskStatus),
		field.String("billing_metric").
			MaxLen(32).
			Optional().
			Nillable(),
		field.Float("reserved_cost").
			Default(0).
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Float("actual_cost").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Float("rate_multiplier").
			Default(1).
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}),
		field.String("billing_currency").
			MaxLen(8).
			Default("USD"),
		field.JSON("request_params", json.RawMessage{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("upstream_usage", json.RawMessage{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("result_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("对客户端展示的结果视频链接（转存后为自有存储链接，否则为上游直链）"),
		field.String("result_storage_key").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("自有对象存储的对象键（转存成功时写入）"),
		field.Int("poll_attempts").
			Default(0),
		field.Time("expires_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("settled_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("error_message").
			Optional().
			Nillable(),
	}
}

func (MediaGenerationTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "status"),
		index.Fields("subscription_id"),
		index.Fields("status", "expires_at"),
		index.Fields("created_at"),
	}
}

func validateMediaTaskStatus(status string) error {
	switch status {
	case "pending", "in_progress", "completed", "failed", "expired", "cancelled":
		return nil
	default:
		return fmt.Errorf("invalid media task status: %s", status)
	}
}

// DefaultMediaTaskTTL 是任务默认过期时间（用于 expires_at 兜底）。
const DefaultMediaTaskTTL = 30 * time.Minute
