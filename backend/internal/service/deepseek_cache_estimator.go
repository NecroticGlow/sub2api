package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	SettingKeyDeepSeekCacheEstimate        = "deepseek_cache_estimate"
	deepSeekCacheEstimateContextKey        = "deepseek_cache_estimate_candidate"
	deepSeekCacheEstimateDefaultLRU        = 16
	deepSeekCacheEstimateDefaultConfidence = 85
)

type deepSeekCacheEstimateSettings struct {
	Enabled           bool      `json:"enabled"`
	AccountGroups     [][]int64 `json:"account_groups"`
	TTLSeconds        int       `json:"ttl_seconds"`
	BlockBytes        int       `json:"block_bytes"`
	MaxFingerprints   int       `json:"max_fingerprints"`
	ConfidencePercent int       `json:"confidence_percent"`
}

type deepSeekCacheFingerprint struct {
	blocks [][32]byte
	bytes  int
	at     time.Time
}

type deepSeekCacheCandidate struct {
	commonBytes       int
	currentBytes      int
	confidencePercent int
}

type deepSeekCacheEstimator struct {
	settings *SettingService
	mu       sync.Mutex
	loadedAt time.Time
	config   deepSeekCacheEstimateSettings
	states   map[string][]deepSeekCacheFingerprint
}

func newDeepSeekCacheEstimator(settings *SettingService) *deepSeekCacheEstimator {
	return &deepSeekCacheEstimator{settings: settings, states: make(map[string][]deepSeekCacheFingerprint)}
}

func (e *deepSeekCacheEstimator) loadConfig(ctx context.Context) deepSeekCacheEstimateSettings {
	e.mu.Lock()
	defer e.mu.Unlock()
	if time.Since(e.loadedAt) < 15*time.Second {
		return e.config
	}
	cfg := deepSeekCacheEstimateSettings{
		TTLSeconds: 900, BlockBytes: 256, MaxFingerprints: deepSeekCacheEstimateDefaultLRU,
		ConfidencePercent: deepSeekCacheEstimateDefaultConfidence,
	}
	if e.settings != nil && e.settings.settingRepo != nil {
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), gatewayForwardingDBTimeout)
		raw, err := e.settings.settingRepo.GetValue(dbCtx, SettingKeyDeepSeekCacheEstimate)
		cancel()
		if err == nil {
			if strings.HasPrefix(strings.TrimSpace(raw), "{") {
				_ = json.Unmarshal([]byte(raw), &cfg)
			} else {
				cfg.Enabled, _ = strconv.ParseBool(strings.TrimSpace(raw))
			}
		}
	}
	if cfg.TTLSeconds < 30 || cfg.TTLSeconds > 3600 {
		cfg.TTLSeconds = 900
	}
	if cfg.BlockBytes < 64 || cfg.BlockBytes > 4096 {
		cfg.BlockBytes = 256
	}
	if cfg.MaxFingerprints < 2 || cfg.MaxFingerprints > 128 {
		cfg.MaxFingerprints = deepSeekCacheEstimateDefaultLRU
	}
	if cfg.ConfidencePercent < 1 || cfg.ConfidencePercent > 100 {
		cfg.ConfidencePercent = deepSeekCacheEstimateDefaultConfidence
	}
	e.config, e.loadedAt = cfg, time.Now()
	return cfg
}

func ollamaCloudAccount(account *Account) bool {
	if account == nil {
		return false
	}
	u, err := url.Parse(strings.TrimSpace(account.GetCredential("base_url")))
	if err != nil {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	return host == "ollama.com" || host == "www.ollama.com"
}

func deepSeekCacheGroup(accountID int64, groups [][]int64) string {
	for i, group := range groups {
		for _, id := range group {
			if id == accountID {
				return "group:" + strconv.Itoa(i)
			}
		}
	}
	return "account:" + strconv.FormatInt(accountID, 10)
}

func canonicalDeepSeekPrompt(body []byte) []byte {
	root := gjson.ParseBytes(body)
	prompt := struct {
		Instructions json.RawMessage `json:"instructions,omitempty"`
		Tools        json.RawMessage `json:"tools,omitempty"`
		Input        json.RawMessage `json:"input,omitempty"`
		Messages     json.RawMessage `json:"messages,omitempty"`
	}{}
	for key, target := range map[string]*json.RawMessage{
		"instructions": &prompt.Instructions,
		"tools":        &prompt.Tools,
		"input":        &prompt.Input,
		"messages":     &prompt.Messages,
	} {
		if value := root.Get(key); value.Exists() && gjson.Valid(value.Raw) {
			*target = append((*target)[:0], value.Raw...)
		}
	}
	canonical, _ := json.Marshal(prompt)
	return canonical
}

func fingerprintDeepSeekPrompt(body []byte, blockBytes int) deepSeekCacheFingerprint {
	canonical := canonicalDeepSeekPrompt(body)
	blocks := make([][32]byte, 0, (len(canonical)+blockBytes-1)/blockBytes)
	for start := 0; start < len(canonical); start += blockBytes {
		end := min(start+blockBytes, len(canonical))
		blocks = append(blocks, sha256.Sum256(canonical[start:end]))
	}
	return deepSeekCacheFingerprint{blocks: blocks, bytes: len(canonical), at: time.Now()}
}

func (e *deepSeekCacheEstimator) prepare(ctx context.Context, c *gin.Context, account *Account, model string, body []byte) {
	if e == nil || c == nil || account == nil || !strings.HasPrefix(strings.ToLower(model), "deepseek-") || !ollamaCloudAccount(account) {
		return
	}
	cfg := e.loadConfig(ctx)
	if !cfg.Enabled {
		return
	}
	current := fingerprintDeepSeekPrompt(body, cfg.BlockBytes)
	if current.bytes == 0 {
		return
	}
	key := deepSeekCacheGroup(account.ID, cfg.AccountGroups) + "\x00" + strings.ToLower(model)
	now := time.Now()
	cutoff := now.Add(-time.Duration(cfg.TTLSeconds) * time.Second)
	e.mu.Lock()
	previous := e.states[key]
	kept := make([]deepSeekCacheFingerprint, 0, min(len(previous)+1, cfg.MaxFingerprints))
	bestCommonBytes := 0
	for i := len(previous) - 1; i >= 0; i-- {
		candidate := previous[i]
		if candidate.at.Before(cutoff) {
			continue
		}
		commonBlocks := 0
		for commonBlocks < len(candidate.blocks) && commonBlocks < len(current.blocks) &&
			candidate.blocks[commonBlocks] == current.blocks[commonBlocks] {
			commonBlocks++
		}
		commonBytes := commonBlocks * cfg.BlockBytes
		if commonBlocks == len(current.blocks) {
			commonBytes = current.bytes
		}
		if commonBytes > bestCommonBytes {
			bestCommonBytes = commonBytes
		}
		kept = append(kept, candidate)
		if len(kept) >= cfg.MaxFingerprints-1 {
			break
		}
	}
	// kept was collected newest-first; order is irrelevant for matching, and
	// appending current keeps it at the newest end for the next LRU scan.
	for left, right := 0, len(kept)-1; left < right; left, right = left+1, right-1 {
		kept[left], kept[right] = kept[right], kept[left]
	}
	e.states[key] = append(kept, current)
	e.mu.Unlock()
	if bestCommonBytes > 0 {
		c.Set(deepSeekCacheEstimateContextKey, deepSeekCacheCandidate{
			commonBytes: bestCommonBytes, currentBytes: current.bytes,
			confidencePercent: cfg.ConfidencePercent,
		})
	}
}

func applyDeepSeekCacheEstimate(c *gin.Context, body []byte) ([]byte, int) {
	if c == nil {
		return body, 0
	}
	raw, ok := c.Get(deepSeekCacheEstimateContextKey)
	if !ok {
		return body, 0
	}
	candidate, ok := raw.(deepSeekCacheCandidate)
	if !ok || candidate.currentBytes <= 0 {
		return body, 0
	}
	usagePath := "usage"
	input := gjson.GetBytes(body, usagePath+".input_tokens").Int()
	detailsPath := "input_tokens_details"
	if input <= 0 {
		input = gjson.GetBytes(body, usagePath+".prompt_tokens").Int()
		detailsPath = "prompt_tokens_details"
	}
	if input <= 0 {
		usagePath = "response.usage"
		input = gjson.GetBytes(body, usagePath+".input_tokens").Int()
		detailsPath = "input_tokens_details"
		if input <= 0 {
			input = gjson.GetBytes(body, usagePath+".prompt_tokens").Int()
			detailsPath = "prompt_tokens_details"
		}
	}
	if input <= 0 {
		return body, 0
	}
	estimated := int(input) * candidate.commonBytes / candidate.currentBytes
	estimated = max(0, min(estimated-128, int(input)-1))
	confidencePercent := candidate.confidencePercent
	if confidencePercent < 1 || confidencePercent > 100 {
		confidencePercent = deepSeekCacheEstimateDefaultConfidence
	}
	estimated = estimated * confidencePercent / 100
	if estimated <= 0 {
		return body, 0
	}
	updated, err := sjson.SetBytes(body, usagePath+"."+detailsPath+".cached_tokens", estimated)
	if err != nil {
		return body, 0
	}
	updated, err = sjson.SetBytes(updated, usagePath+"."+detailsPath+".cached_tokens_estimated", true)
	if err != nil {
		return body, 0
	}
	return updated, estimated
}
