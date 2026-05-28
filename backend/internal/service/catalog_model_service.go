package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

//go:embed catalog_models.json
var catalogSeedJSON []byte

// CatalogModelService handles business logic for the model catalog.
type CatalogModelService struct {
	repo CatalogModelRepository
}

// NewCatalogModelService creates a new CatalogModelService.
func NewCatalogModelService(repo CatalogModelRepository) *CatalogModelService {
	return &CatalogModelService{repo: repo}
}

// SeedIfEmpty inserts the bundled catalog_models.json into the DB if the table is empty.
// Non-fatal: logs on parse/insert error so startup is not blocked.
func (s *CatalogModelService) SeedIfEmpty(ctx context.Context) error {
	count, err := s.repo.CountAll(ctx)
	if err != nil {
		return fmt.Errorf("catalog seed count: %w", err)
	}
	if count > 0 {
		return nil
	}

	models, err := parseSeedJSON(catalogSeedJSON)
	if err != nil {
		log.Printf("[CatalogModel] Warning: seed JSON parse error: %v", err)
		return nil
	}

	if err := s.repo.BulkUpsert(ctx, models); err != nil {
		log.Printf("[CatalogModel] Warning: seed insert error: %v", err)
		return nil
	}
	log.Printf("[CatalogModel] Seeded %d models into catalog_models table", len(models))
	return nil
}

// Reseed re-upserts all bundled catalog data (idempotent — uses OnConflict).
func (s *CatalogModelService) Reseed(ctx context.Context) (int, error) {
	models, err := parseSeedJSON(catalogSeedJSON)
	if err != nil {
		return 0, fmt.Errorf("seed JSON parse: %w", err)
	}
	if err := s.repo.BulkUpsert(ctx, models); err != nil {
		return 0, fmt.Errorf("seed upsert: %w", err)
	}
	return len(models), nil
}

// List returns a paginated list of catalog models with optional filters.
func (s *CatalogModelService) List(ctx context.Context, params pagination.PaginationParams, filters CatalogModelListFilters) ([]CatalogModel, *pagination.PaginationResult, error) {
	return s.repo.List(ctx, params, filters)
}

// GetByID returns a single catalog model by its ent int64 ID.
func (s *CatalogModelService) GetByID(ctx context.Context, id int64) (*CatalogModel, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByModelID returns a single catalog model by its string model_id (e.g. "openai/gpt-4o").
func (s *CatalogModelService) GetByModelID(ctx context.Context, modelID string) (*CatalogModel, error) {
	return s.repo.GetByModelID(ctx, modelID)
}

// ListEnabled returns all enabled models for the public catalog.
func (s *CatalogModelService) ListEnabled(ctx context.Context) ([]CatalogModel, error) {
	return s.repo.ListEnabled(ctx)
}

// Update applies full-field update to the catalog model identified by int64 ID.
func (s *CatalogModelService) Update(ctx context.Context, id int64, input UpdateCatalogModelInput) (*CatalogModel, error) {
	return s.repo.Update(ctx, id, input)
}

// Toggle flips the is_enabled flag for the catalog model identified by int64 ID.
func (s *CatalogModelService) Toggle(ctx context.Context, id int64) (*CatalogModel, error) {
	return s.repo.Toggle(ctx, id)
}

// seedPricingJSON holds pricing fields from the JSON file.
type seedPricingJSON struct {
	Input          string `json:"input"`
	Output         string `json:"output"`
	InputCacheRead string `json:"input_cache_read"`
	InputCacheWrite string `json:"input_cache_write"`
}

// seedCapabilitiesJSON holds capability flags from the JSON file.
type seedCapabilitiesJSON struct {
	Vision          bool `json:"vision"`
	FunctionCalling bool `json:"function_calling"`
	Reasoning       bool `json:"reasoning"`
	PromptCaching   bool `json:"prompt_caching"`
	WebSearch       bool `json:"web_search"`
	WebFetch        bool `json:"web_fetch"`
	AudioInput      bool `json:"audio_input"`
	VideoInput      bool `json:"video_input"`
	PDFInput        bool `json:"pdf_input"`
	Streaming       bool `json:"streaming"`
	JSONMode        bool `json:"json_mode"`
	FileUpload      bool `json:"file_upload"`
}

// seedModelJSON is the shape of each record in catalog_models.json.
type seedModelJSON struct {
	ID                     string               `json:"id"`
	DisplayName            string               `json:"display_name"`
	Icon                   string               `json:"icon"`
	Mode                   string               `json:"mode"`
	ContextWindow          int64                `json:"context_window"`
	MaxOutputTokens        int64                `json:"max_output_tokens"`
	Pricing                seedPricingJSON      `json:"pricing"`
	Capabilities           seedCapabilitiesJSON `json:"capabilities"`
	SupportedProviderTypes []string             `json:"supported_provider_types"`
	SupportedProtocols     []string             `json:"supported_protocols"`
	ReleasedAt             string               `json:"released_at"`
}

func parseSeedJSON(data []byte) ([]CatalogModel, error) {
	var raw []seedModelJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	models := make([]CatalogModel, 0, len(raw))
	for _, r := range raw {
		models = append(models, seedToDomain(r))
	}
	return models, nil
}

func seedToDomain(r seedModelJSON) CatalogModel {
	vendor := vendorFromModelID(r.ID)
	features := buildFeatures(r.Capabilities)

	var cacheReadPrice *float64
	var cacheWritePrice *float64
	if r.Pricing.InputCacheRead != "" {
		v := parsePrice(r.Pricing.InputCacheRead)
		cacheReadPrice = &v
	}
	if r.Pricing.InputCacheWrite != "" {
		v := parsePrice(r.Pricing.InputCacheWrite)
		cacheWritePrice = &v
	}

	return CatalogModel{
		ModelID:          r.ID,
		Name:             r.DisplayName,
		Vendor:           vendor,
		Category:         r.Mode,
		IconURL:          r.Icon,
		ContextWindow:    r.ContextWindow,
		MaxOutputTokens:  r.MaxOutputTokens,
		InputModalities:  buildInputModalities(r.Capabilities),
		OutputModalities: []string{"text"},
		Features:         features,
		InputPrice:       parsePrice(r.Pricing.Input),
		OutputPrice:      parsePrice(r.Pricing.Output),
		CacheReadPrice:   cacheReadPrice,
		CacheWritePrice:  cacheWritePrice,
		Currency:         "USD",
		IsEnabled:        isDefaultEnabledVendor(vendor),
		Tags:             []string{},
		Description:      "",
	}
}

func isDefaultEnabledVendor(vendor string) bool {
	switch vendor {
	case "OpenAI", "Anthropic":
		return true
	default:
		return false
	}
}

func buildFeatures(caps seedCapabilitiesJSON) []string {
	capMap := []struct {
		enabled bool
		name    string
	}{
		{caps.Streaming, "streaming"},
		{caps.FunctionCalling, "function_calling"},
		{caps.Vision, "vision"},
		{caps.JSONMode, "json_mode"},
		{caps.FileUpload, "file_upload"},
		{caps.Reasoning, "reasoning"},
		{caps.PromptCaching, "prompt_caching"},
		{caps.WebSearch, "web_search"},
		{caps.WebFetch, "web_fetch"},
		{caps.AudioInput, "audio_input"},
		{caps.VideoInput, "video_input"},
		{caps.PDFInput, "pdf_input"},
	}
	features := make([]string, 0, len(capMap))
	for _, c := range capMap {
		if c.enabled {
			features = append(features, c.name)
		}
	}
	return features
}

func buildInputModalities(caps seedCapabilitiesJSON) []string {
	modalities := []string{"text"}
	if caps.Vision || caps.VideoInput {
		modalities = append(modalities, "image")
	}
	if caps.AudioInput {
		modalities = append(modalities, "audio")
	}
	if caps.PDFInput {
		modalities = append(modalities, "pdf")
	}
	return modalities
}

func vendorFromModelID(id string) string {
	prefixes := []struct{ prefix, name string }{
		{"openai/", "OpenAI"},
		{"anthropic/", "Anthropic"},
		{"google/", "Google"},
		{"meta/", "Meta"},
		{"mistral/", "Mistral"},
		{"xai/", "xAI"},
		{"deepseek/", "DeepSeek"},
		{"qwen/", "Alibaba"},
		{"bailian/", "Alibaba"},
		{"cohere/", "Cohere"},
		{"perplexity/", "Perplexity"},
		{"z-ai/", "Zhipu"},
		{"zhipu/", "Zhipu"},
	}
	for _, p := range prefixes {
		if len(id) > len(p.prefix) && id[:len(p.prefix)] == p.prefix {
			return p.name
		}
	}
	return "Other"
}

func parsePrice(s string) float64 {
	if s == "" {
		return 0
	}
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}
