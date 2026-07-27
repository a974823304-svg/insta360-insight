package source

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenProvider 提供调用平台接口所需的 access_token。
// 取到 token 失败(无凭证 / 平台报错)时返回 error,
// adapter 捕获后返回 ErrNotImplemented,触发 FallbackDataSource 回退到 MockAdapter。
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

// StaticToken 直接使用配置里已授权的 token(最简单,推荐先用这个)。
type StaticToken struct {
	token string
}

func NewStaticToken(t string) *StaticToken { return &StaticToken{token: t} }

func (s *StaticToken) Token(_ context.Context) (string, error) {
	if s.token == "" {
		return "", fmt.Errorf("no static access_token configured")
	}
	return s.token, nil
}

// ClientToken 通过 client_key/secret 走 client_token 流程(无需用户授权)。
// 目前仅抖音开放平台提供应用级 client_token;B站/小红书无此模式,请勿使用。
type ClientToken struct {
	mu       sync.Mutex
	baseURL  string
	key      string
	secret   string
	cached   string
	expireAt time.Time
	httpDo   func(req *http.Request) (*http.Response, error)
}

// NewClientToken 构造抖音 client_token provider。httpDo 可注入用于测试。
func NewClientToken(baseURL, key, secret string) *ClientToken {
	return &ClientToken{
		baseURL: strings.TrimRight(baseURL, "/"),
		key:     key,
		secret:  secret,
		httpDo:  http.DefaultClient.Do,
	}
}

type clientTokenResp struct {
	Message string `json:"message"`
	Data    struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	} `json:"data"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (c *ClientToken) Token(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cached != "" && time.Now().Before(c.expireAt.Add(-5*time.Minute)) {
		return c.cached, nil
	}
	if c.key == "" || c.secret == "" {
		return "", fmt.Errorf("client_key/secret not configured")
	}
	u := c.baseURL + "/oauth/client_token/"
	q := url.Values{}
	q.Set("client_key", c.key)
	q.Set("client_secret", c.secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u+"?"+q.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpDo(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var r clientTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	if r.Data.AccessToken == "" {
		return "", fmt.Errorf("get client_token failed: errcode=%d %s %s", r.ErrCode, r.ErrMsg, r.Message)
	}
	c.cached = r.Data.AccessToken
	// 抖音 client_token 有效期 2h;留 5 分钟缓冲。
	c.expireAt = time.Now().Add(time.Duration(r.Data.ExpiresIn) * time.Second)
	return c.cached, nil
}

// resolveTokenProvider 根据平台配置解析出一个 TokenProvider。
//  1. 配了 ACCESS_TOKEN -> StaticToken
//  2. 配了 CLIENT_KEY/SECRET 且 allowClientToken -> 抖音 client_token
//  3. 都没有 -> 返回 nil(adapter 所有方法将返回 ErrNotImplemented,触发回退)
func resolveTokenProvider(cfg PlatformConfig, allowClientToken bool) TokenProvider {
	if cfg.AccessToken != "" {
		return NewStaticToken(cfg.AccessToken)
	}
	if allowClientToken && cfg.ClientKey != "" && cfg.ClientSecret != "" {
		return NewClientToken(cfg.BaseURL, cfg.ClientKey, cfg.ClientSecret)
	}
	return nil
}

// md5Base64 计算字符串的 MD5 并做 base64 编码(用于 B站签名 Content-MD5)。
func md5Base64(s string) string {
	sum := md5.Sum([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}
