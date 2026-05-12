package config

import "github.com/dtapps/weibo-go/types"

// DefaultConfig 返回默认配置
func DefaultConfig() *types.WeiboConfig {
	return &types.WeiboConfig{
		Enabled:       new(true),
		WSEndpoint:    types.DefaultWSEndpoint,
		TokenEndpoint: types.DefaultTokenEndpoint,
		DMPolicy:      "open",
		ChunkMode:     "raw",
	}
}
