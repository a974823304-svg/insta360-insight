package source

import "os"

// PlatformConfig 单个平台的接入凭证与端点配置。
// 全部从环境变量读取;缺省时对应 adapter 的方法返回 ErrNotImplemented,
// 由 FallbackDataSource 自动回退到 MockAdapter(看板永远有数据、不崩溃)。
//
// 凭证优先级(在 adapter 内解析):
//  1. ACCESS_TOKEN  —— 已通过 OAuth2 授权拿到的 token(最简单,推荐先用这个)
//  2. CLIENT_KEY/SECRET —— 走 client_token 流程(仅抖音开放平台支持无需用户授权的 client_token)
//
// 注意:B站 / 小红书没有"无需用户授权的应用级 token",必须先走授权码流程拿到
// 用户/商家 access_token 再填到 *_ACCESS_TOKEN;只配 client_id/secret 不足以拉数。
type PlatformConfig struct {
	ClientKey    string
	ClientSecret string
	AccessToken  string // 已授权的 access_token / 聚合 API key(可选,优先于 client_token 流程)
	UserID       string // 平台内的用户/达人 ID(如小红书蒲公英 userId)
	BaseURL      string
}

// Config 三平台接入配置聚合。
type Config struct {
	Douyin      PlatformConfig
	Bilibili    PlatformConfig
	Xiaohongshu PlatformConfig
}

// LoadConfig 从环境变量装载配置。
//
// 变量约定:
//   DOUYIN_CLIENT_KEY / DOUYIN_CLIENT_SECRET / DOUYIN_ACCESS_TOKEN / DOUYIN_BASE_URL
//   BILIBILI_CLIENT_ID / BILIBILI_CLIENT_SECRET / BILIBILI_ACCESS_TOKEN / BILIBILI_BASE_URL
//   XHS_APP_ID / XHS_APP_SECRET / XHS_ACCESS_TOKEN / XHS_BASE_URL
func LoadConfig() Config {
	return Config{
		Douyin: PlatformConfig{
			ClientKey:    os.Getenv("DOUYIN_CLIENT_KEY"),
			ClientSecret: os.Getenv("DOUYIN_CLIENT_SECRET"),
			AccessToken:  os.Getenv("DOUYIN_ACCESS_TOKEN"),
			BaseURL:      orDefault(os.Getenv("DOUYIN_BASE_URL"), "https://open.douyin.com"),
		},
		Bilibili: PlatformConfig{
			ClientKey:    os.Getenv("BILIBILI_CLIENT_ID"),
			ClientSecret: os.Getenv("BILIBILI_CLIENT_SECRET"),
			AccessToken:  os.Getenv("BILIBILI_ACCESS_TOKEN"),
			BaseURL:      orDefault(os.Getenv("BILIBILI_BASE_URL"), "https://member.bilibili.com"),
		},
		Xiaohongshu: PlatformConfig{
			ClientKey:    os.Getenv("XHS_APP_ID"),
			ClientSecret: os.Getenv("XHS_APP_SECRET"),
			AccessToken:  os.Getenv("XHS_ACCESS_TOKEN"),
			UserID:       os.Getenv("XHS_USER_ID"),
			BaseURL:      orDefault(os.Getenv("XHS_BASE_URL"), "https://ark.xiaohongshu.com"),
		},
	}
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
