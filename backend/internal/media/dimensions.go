package media

import (
	"math"
	"strconv"
	"strings"
)

// resolveVideoDimensions 由分辨率档位与宽高比推导标准输出宽高。
//
// 用于预扣估算：文生/图生视频请求常只给 resolution + aspect_ratio 而不给
// 精确宽高，此时按标准档位补齐，保证 token 估算（及预扣）非零。
// 短边取分辨率档位基准（480/720/1080/2160），长边按宽高比推导并取偶数。
// 未知分辨率返回 ok=false。
func resolveVideoDimensions(resolution, aspectRatio string) (w, h int, ok bool) {
	short, ok := resolutionShortSide(resolution)
	if !ok {
		return 0, 0, false
	}
	rw, rh := parseAspectRatio(aspectRatio)

	hi, lo := rw, rh
	if hi < lo {
		hi, lo = lo, hi
	}
	long := roundToEven(float64(short) * float64(hi) / float64(lo))

	// rw >= rh：横向或方形，长边为宽；否则竖向，长边为高。
	if rw >= rh {
		return long, short, true
	}
	return short, long, true
}

// resolutionShortSide 把分辨率档位映射为短边像素基准。
func resolutionShortSide(resolution string) (int, bool) {
	r := strings.ToLower(strings.TrimSpace(resolution))
	r = strings.TrimSuffix(r, "p")
	switch r {
	case "480":
		return 480, true
	case "720":
		return 720, true
	case "1080":
		return 1080, true
	case "1440", "2k":
		return 1440, true
	case "2160", "4k":
		return 2160, true
	}
	return 0, false
}

// parseAspectRatio 解析 "w:h" 宽高比；缺省或非法时返回 16:9。
func parseAspectRatio(aspectRatio string) (int, int) {
	parts := strings.Split(strings.TrimSpace(aspectRatio), ":")
	if len(parts) == 2 {
		rw, errW := strconv.Atoi(strings.TrimSpace(parts[0]))
		rh, errH := strconv.Atoi(strings.TrimSpace(parts[1]))
		if errW == nil && errH == nil && rw > 0 && rh > 0 {
			return rw, rh
		}
	}
	return 16, 9
}

// roundToEven 四舍五入到最近整数，若为奇数则进位为偶数（视频宽高通常为偶数）。
func roundToEven(f float64) int {
	r := int(math.Round(f))
	if r%2 == 1 {
		r++
	}
	return r
}
