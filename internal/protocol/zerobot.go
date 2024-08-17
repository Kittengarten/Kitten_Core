package protocol

import (
	"github.com/Kittengarten/KittenCore/kitten"

	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

func getDriver(forward bool, url, accessToken string) zero.Driver {
	if forward {
		return &driver.WSClient{
			// OneBot 正向 WS 默认使用 6700 端口
			Url:         url,
			AccessToken: accessToken,
		}
	}
	return &driver.WSServer{
		// OneBot 反向 WS 默认使用 6700 端口
		Url:         url,
		AccessToken: accessToken,
	}
}

// Runbot 启动机器人
func RunBot(forward bool) {
	var (
		config     = kitten.MainConfig()
		superUsers = make([]int64, len(config.SuperUsers), len(config.SuperUsers))
	)
	for i, u := range config.SuperUsers {
		superUsers[i] = u.Int()
	}
	zero.RunAndBlock(&zero.Config{
		NickName:      config.NickName,
		CommandPrefix: config.CommandPrefix,
		SuperUsers:    superUsers,
		Driver: []zero.Driver{
			getDriver(forward,
				config.WebSocket.URL,
				config.WebSocket.AccessToken),
		},
	}, process.GlobalInitMutex.Unlock)
}
