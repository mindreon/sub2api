package service

func ResolveMediaRateMultiplier(apiKey *APIKey, effectiveGroupMultiplier float64) float64 {
	if apiKey != nil && apiKey.Group != nil && apiKey.Group.MediaRateIndependent {
		if apiKey.Group.MediaRateMultiplier < 0 {
			return 0
		}
		return apiKey.Group.MediaRateMultiplier
	}
	return effectiveGroupMultiplier
}

// IsMediaPlatform 报告平台是否用于多模态异步上游（火山 / OpenRouter）。
func IsMediaPlatform(platform string) bool {
	switch platform {
	case PlatformVolcengine, PlatformOpenRouter:
		return true
	default:
		return false
	}
}
