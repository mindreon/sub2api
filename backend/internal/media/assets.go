package media

import (
	"context"
	"strconv"
)

// AssetStore 把上游临时资源（如火山返回的、24h 过期的 video_url）转存到平台
// 自有对象存储，并按需为已存储对象签发可访问链接。
//
// media 包只定义端口；具体实现（复用 S3 兼容存储）在组合根注入，因此本包
// 对上游存储实现零 import。未配置存储时组合根注入 nil，结算逻辑自动降级为
// 保留上游直链。
type AssetStore interface {
	// Rehost 下载 srcURL 并以 key 存入自有存储，返回可访问链接。
	Rehost(ctx context.Context, key, srcURL, contentType string) (url string, err error)
	// PresignedURL 为已存储对象签发新的临时可访问链接（无公共域名直链时使用）。
	PresignedURL(ctx context.Context, key string) (string, error)
}

// mediaAssetKey 生成对象存储键：media/{userID}/{taskID}.mp4。
func mediaAssetKey(userID int64, taskID string) string {
	return "media/" + strconv.FormatInt(userID, 10) + "/" + taskID + ".mp4"
}
