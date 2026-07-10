package service

// AllowsMediaGeneration 判断分组是否允许调用 /v1/video/generations 等多模态异步接口。
func AllowsMediaGeneration(group *Group) bool {
	if group == nil {
		return false
	}
	if !IsMediaPlatform(group.Platform) {
		return false
	}
	return group.AllowMediaGeneration
}
