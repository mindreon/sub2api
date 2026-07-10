package media

import "strings"

// Backend 标识应由哪家上游厂商处理该模型。
type Backend string

const (
	BackendVolcengine Backend = "volcengine"
	BackendOpenRouter Backend = "openrouter"
)

// RouteModel 根据模型名解析应使用的厂商后端。
// 第二个返回值 false 表示不是已知的 Seedance 多模态模型。
func RouteModel(model string) (Backend, bool) {
	m := normalizeModelKey(model)
	if m == "" {
		return "", false
	}
	if strings.HasPrefix(m, "bytedance/") {
		if isSeedanceFamily(m) {
			return BackendOpenRouter, true
		}
		return "", false
	}
	if isSeedanceFamily(m) {
		return BackendVolcengine, true
	}
	return "", false
}

// normalizeModelKey 统一模型名大小写与前缀，便于路由与定价查表。
func normalizeModelKey(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}

func isSeedanceFamily(model string) bool {
	m := model
	m = strings.TrimPrefix(m, "doubao-")
	m = strings.TrimPrefix(m, "bytedance/")
	m = strings.TrimPrefix(m, "volcengine/")
	return strings.Contains(m, "seedance")
}
