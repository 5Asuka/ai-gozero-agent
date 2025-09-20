package config

import "github.com/zeromicro/go-zero/rest"

// 定义接口 将配置文件传递给handler
type Config struct {
	rest.RestConf
	OpenAI struct {
		ApiKey      string
		BaseUrl     string
		Model       string
		MaxTokens   int
		Temperature float32
	}
}
