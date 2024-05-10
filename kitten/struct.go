package kitten

import "github.com/Kittengarten/KittenCore/kitten/core"

type (
	// 来自 Bot 的配置文件的数据集
	config struct {
		NickName      []string        `yaml:"nickname"`      // 昵称
		SelfID        QQ              `yaml:"selfid"`        // Bot 自身 ID
		SuperUsers    []QQ            `yaml:"superusers"`    // 亲妈账号
		CommandPrefix string          `yaml:"commandprefix"` // 指令前缀
		Path          core.Path       `yaml:"path"`          // 资源文件路径
		WebSocket     WebSocketConfig `yaml:"websocket"`     // WebSocket 配置
		Log           LogConfig       `yaml:"log"`           // 日志配置
		WebUI         WebUIConfig     `yaml:"webui"`         // WebUI 配置
	}

	// WebSocketConfig 是一个 WebSocket 链接的配置
	WebSocketConfig struct {
		URL         string `yaml:"url"`         // WebSocket 链接
		AccessToken string `yaml:"accesstoken"` // WebSocket 密钥
	}

	// WebUIConfig 是一个 WebUI 的配置
	WebUIConfig struct {
		URL string `yaml:"url"` // WebUI 链接
	}

	// LogConfig 是一个日志的配置
	LogConfig struct {
		Level      string `yaml:"level"`      // 日志等级
		Path       string `yaml:"path"`       // 日志路径
		MaxSize    int    `yaml:"maxsize"`    // 文件大小限制，单位 MB
		MaxBackups int    `yaml:"maxbackups"` // 最大保留日志文件数量
		MaxAge     int    `yaml:"expire"`     // 日志文件的过期天数，大于该天数前的日志文件会被清理。设置为 -1 可以禁用。
	}
)
