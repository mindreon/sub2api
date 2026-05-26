package dto

// PublicModel is the display-facing model DTO served to gasboard and external consumers.
// Only fields needed for the public catalog are included — internal scheduling and cost
// details are intentionally omitted.
type PublicModel struct {
	ID                     string            `json:"id"`
	DisplayName            string            `json:"display_name"`
	Icon                   string            `json:"icon"`
	Mode                   string            `json:"mode"`
	ContextWindow          int               `json:"context_window"`
	MaxOutputTokens        int               `json:"max_output_tokens"`
	Pricing                map[string]string `json:"pricing"`
	Capabilities           map[string]bool   `json:"capabilities"`
	SupportedProviderTypes []string          `json:"supported_provider_types"`
	SupportedProtocols     []string          `json:"supported_protocols,omitempty"`
	ReleasedAt             string            `json:"released_at"`
}

// PublicModelsListResponse wraps the model list for the GET /api/public/models endpoint.
type PublicModelsListResponse struct {
	Success bool          `json:"success"`
	Data    []PublicModel `json:"data"`
	Total   int           `json:"total"`
}

// PublicModelDetailResponse wraps a single model for the GET /api/public/models/:id endpoint.
type PublicModelDetailResponse struct {
	Success bool        `json:"success"`
	Data    PublicModel `json:"data"`
}
