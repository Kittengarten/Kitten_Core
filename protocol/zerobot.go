package protocol

import (
	"github.com/Kittengarten/KittenCore/kitten"

	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

func RunBot() {
	config := kitten.MainConfig()
	zero.RunAndBlock(&zero.Config{
		NickName:      config.NickName,
		CommandPrefix: config.CommandPrefix,
		SuperUsers:    config.SuperUsers,
		Driver: []zero.Driver{
			&driver.WSClient{
				// OneBot 正向 WS 默认使用 6700 端口
				Url:         config.WebSocket.URL,
				AccessToken: config.WebSocket.AccessToken,
			},
		},
	}, process.GlobalInitMutex.Unlock)
}
