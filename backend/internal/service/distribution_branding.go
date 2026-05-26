package service

import (
	"context"
	"net"
	"net/url"
	"strings"
)

type DistributionBrandingResolver interface {
	GetByBrandHost(ctx context.Context, host string) (*DistributionOrganization, error)
}

type publicSettingsRequestMeta struct {
	Host   string
	Scheme string
}

type publicSettingsRequestMetaKey struct{}

func WithPublicSettingsRequestMeta(ctx context.Context, host, scheme string) context.Context {
	meta := publicSettingsRequestMeta{
		Host:   NormalizeDistributionBrandHost(host),
		Scheme: normalizePublicSettingsRequestScheme(scheme),
	}
	return context.WithValue(ctx, publicSettingsRequestMetaKey{}, meta)
}

func publicSettingsRequestMetaFromContext(ctx context.Context) publicSettingsRequestMeta {
	meta, _ := ctx.Value(publicSettingsRequestMetaKey{}).(publicSettingsRequestMeta)
	meta.Host = NormalizeDistributionBrandHost(meta.Host)
	meta.Scheme = normalizePublicSettingsRequestScheme(meta.Scheme)
	return meta
}

func NormalizeDistributionBrandHost(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if idx := strings.Index(value, ","); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}
	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err == nil && parsed.Host != "" {
			value = parsed.Host
		}
	} else if strings.Contains(value, "/") {
		parsed, err := url.Parse("https://" + value)
		if err == nil && parsed.Host != "" {
			value = parsed.Host
		}
	}
	if host, _, err := net.SplitHostPort(value); err == nil && host != "" {
		value = host
	}
	value = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
	return value
}

func ResolveDistributionBrandURL(raw string, scheme string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err == nil && parsed.Host != "" {
			parsed.Path = strings.TrimSpace(parsed.Path)
			return strings.TrimRight(parsed.String(), "/")
		}
	}
	host := NormalizeDistributionBrandHost(value)
	if host == "" {
		return ""
	}
	return normalizePublicSettingsRequestScheme(scheme) + "://" + host
}

func normalizePublicSettingsRequestScheme(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "http":
		return "http"
	default:
		return "https"
	}
}

func distributionBrandString(config map[string]any, key string) string {
	if len(config) == 0 {
		return ""
	}
	value, ok := config[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
