package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"golang.org/x/mod/semver"
)

const defaultClaudeUsageURL = "https://api.anthropic.com/api/oauth/usage"

type claudeUsageService struct {
	usageURL          string
	allowPrivateHosts bool
	httpUpstream      service.HTTPUpstream
}

// NewClaudeUsageFetcher 创建 Claude 用量获取服务
// httpUpstream: 可选，如果提供则支持 TLS 指纹伪装
func NewClaudeUsageFetcher(httpUpstream service.HTTPUpstream) service.ClaudeUsageFetcher {
	return &claudeUsageService{
		usageURL:     defaultClaudeUsageURL,
		httpUpstream: httpUpstream,
	}
}

// FetchUsage 简单版本，不支持 TLS 指纹（向后兼容）
func (s *claudeUsageService) FetchUsage(ctx context.Context, accessToken, proxyURL string) (*service.ClaudeUsageResponse, error) {
	return s.FetchUsageWithOptions(ctx, &service.ClaudeUsageFetchOptions{
		AccessToken: accessToken,
		ProxyURL:    proxyURL,
	})
}

// FetchUsageWithOptions 支持账号代理/TLS；配额 UA 采用支持重置卡的版本下限。
func (s *claudeUsageService) FetchUsageWithOptions(ctx context.Context, opts *service.ClaudeUsageFetchOptions) (*service.ClaudeUsageResponse, error) {
	if opts == nil {
		return nil, fmt.Errorf("options is nil")
	}

	// 创建请求
	usageURL, err := url.Parse(s.usageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Claude usage URL")
	}
	query := usageURL.Query()
	query.Set("cedar_ember", "1")
	query.Set("skip_spend", "1")
	usageURL.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", usageURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头（与抓包一致，但不设置 Accept-Encoding，让 Go 自动处理压缩）
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+opts.AccessToken)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	// 老 Fingerprint 的 UA 会被重置卡接口拒绝，使用配额专用版本下限。
	userAgent := claudeQuotaUserAgent()
	req.Header.Set("User-Agent", userAgent)

	var resp *http.Response

	// 如果有 TLS Profile 且有 HTTPUpstream，使用 DoWithTLS
	if opts.TLSProfile != nil && s.httpUpstream != nil {
		resp, err = s.httpUpstream.DoWithTLS(req, opts.ProxyURL, opts.AccountID, 0, opts.TLSProfile)
		if err != nil {
			return nil, fmt.Errorf("request with TLS fingerprint failed: %w", err)
		}
	} else {
		// 不启用 TLS 指纹，使用普通 HTTP 客户端
		client, err := httpclient.GetClient(httpclient.Options{
			ProxyURL:           opts.ProxyURL,
			Timeout:            30 * time.Second,
			ValidateResolvedIP: true,
			AllowPrivateHosts:  s.allowPrivateHosts,
		})
		if err != nil {
			return nil, fmt.Errorf("create http client failed: %w", err)
		}

		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		msg := fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body))
		return nil, infraerrors.New(http.StatusInternalServerError, "UPSTREAM_ERROR", msg)
	}

	var usageResp service.ClaudeUsageResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&usageResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &usageResp, nil
}

// The banked reset API rejects older Claude Code versions (cli_version), even
// though ordinary usage works. This floor is quota-only: it must not rewrite
// the account's gateway fingerprint or change the global client version.
func claudeQuotaUserAgent() string {
	version := claude.EffectiveCLIVersion()
	if semver.Compare("v"+version, "v2.1.280") < 0 {
		version = "2.1.280"
	}
	return "claude-cli/" + version + " (external, cli)"
}

func (s *claudeUsageService) ResetQuota(ctx context.Context, opts *service.ClaudeUsageFetchOptions, org, grantID, requestID string) (*service.ClaudeResetResult, error) {
	if opts == nil {
		return nil, fmt.Errorf("options is nil")
	}
	body, err := json.Marshal(map[string]string{"program": "cedar_ember", "grant_id": grantID, "request_id": requestID})
	if err != nil {
		return nil, err
	}
	// Fixed official origin; never use a user-supplied account base_url here.
	endpoint := "https://api.anthropic.com/api/organizations/" + url.PathEscape(org) + "/reset_rate_limits"
	// usageURL is only overridden by in-package synthetic HTTP tests.
	if s.allowPrivateHosts {
		u, _ := url.Parse(s.usageURL)
		endpoint = u.Scheme + "://" + u.Host + "/api/organizations/" + url.PathEscape(org) + "/reset_rate_limits"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+opts.AccessToken)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", claudeQuotaUserAgent())
	var resp *http.Response
	if s.httpUpstream != nil {
		resp, err = s.httpUpstream.DoWithTLS(req, opts.ProxyURL, opts.AccountID, 0, opts.TLSProfile)
	} else {
		base, clientErr := httpclient.GetClient(httpclient.Options{ProxyURL: opts.ProxyURL, Timeout: 20 * time.Second, ValidateResolvedIP: true, AllowPrivateHosts: s.allowPrivateHosts})
		if clientErr != nil {
			return nil, infraerrors.New(502, "CLAUDE_RESET_CLIENT_FAILED", "unable to create Claude reset client")
		}
		client := *base
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		resp, err = client.Do(req)
	}
	// Never retry a consuming POST here. Do not log raw bodies or transport
	// errors containing authenticated proxy URLs. Replays use the same UUID.
	if err != nil {
		return nil, infraerrors.New(502, "CLAUDE_RESET_OUTCOME_UNKNOWN", "reset outcome unknown; refresh status and retry only the same grant")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, infraerrors.Newf(502, "CLAUDE_RESET_UPSTREAM_ERROR", "Claude reset returned HTTP %d; refresh before retrying", resp.StatusCode)
	}
	var result service.ClaudeResetResult
	if json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result) != nil {
		return nil, infraerrors.New(502, "CLAUDE_RESET_OUTCOME_UNKNOWN", "invalid reset response; refresh status before retrying")
	}
	switch result.Result {
	case "reset", "already_used", "not_limited", "cooldown", "ineligible", "unavailable":
		return &result, nil
	default:
		return nil, infraerrors.New(502, "CLAUDE_RESET_OUTCOME_UNKNOWN", "unknown reset response; refresh status before retrying")
	}
}
