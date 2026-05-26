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

// CatalogModel holds the schema definition for the public model catalog.
type CatalogModel struct {
	ent.Schema
}

func (CatalogModel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "catalog_models"},
	}
}

func (CatalogModel) Fields() []ent.Field {
	return []ent.Field{
		field.String("model_id").
			Unique().
			NotEmpty().
			MaxLen(200).
			Comment("模型唯一标识，如 openai/gpt-4o"),
		field.String("name").
			MaxLen(200).
			NotEmpty().
			Comment("模型显示名称"),
		field.String("vendor").
			MaxLen(100).
			Default("").
			Comment("厂家名称，如 OpenAI、Anthropic、Google"),
		field.String("category").
			MaxLen(50).
			Default("chat").
			Comment("模型分类: chat, embedding, image, audio, video"),
		field.String("description").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default("").
			Comment("模型描述"),
		field.JSON("tags", []string{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("标签列表"),
		field.String("doc_url").
			Default("").
			Comment("文档链接"),
		field.String("icon_url").
			Default("").
			Comment("图标链接"),
		field.Int64("context_window").
			Default(0).
			Comment("上下文窗口大小（tokens）"),
		field.Int64("max_output_tokens").
			Default(0).
			Comment("最大输出 tokens"),
		field.JSON("input_modalities", []string{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("输入模态: text, image, audio, video"),
		field.JSON("output_modalities", []string{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("输出模态: text, image, audio"),
		field.JSON("features", []string{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("能力特性: streaming, function_calling, vision, json_mode, file_upload"),
		field.Float("input_price").
			Default(0).
			Comment("输入价格（美元/百万 tokens）"),
		field.Float("output_price").
			Default(0).
			Comment("输出价格（美元/百万 tokens）"),
		field.Float("cache_write_price").
			Optional().
			Nillable().
			Comment("缓存写入价格（可选）"),
		field.Float("cache_read_price").
			Optional().
			Nillable().
			Comment("缓存读取价格（可选）"),
		field.String("currency").
			MaxLen(10).
			Default("USD").
			Comment("货币单位"),
		field.Bool("is_enabled").
			Default(true).
			Comment("是否上架"),
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

func (CatalogModel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("model_id").Unique(),
		index.Fields("is_enabled"),
		index.Fields("vendor"),
		index.Fields("category"),
	}
}
