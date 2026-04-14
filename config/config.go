package config

import "github.com/dtapps/weibo-go/types"

// DefaultConfig 返回默认配置
func DefaultConfig() *types.WeiboConfig {
	return &types.WeiboConfig{
		Enabled:       boolPtr(true),
		WSEndpoint:    types.DefaultWSEndpoint,
		TokenEndpoint: types.DefaultTokenEndpoint,
		DMPolicy:      "open",
		ChunkMode:     "raw",
	}
}

func boolPtr(b bool) *bool {
	return &b
}
