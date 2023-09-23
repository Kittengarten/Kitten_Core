// KittenCore 的主函数所在包
package main

import (
	// 内置库
	"runtime/debug"

	// KittenCore 的核心库
	"github.com/Kittengarten/KittenCore/kitten"

	// 核心依赖
	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"go.uber.org/zap"

	// 内部插件
	// _ "github.com/Kittengarten/KittenCore/auth"    // 黑名单控制插件
	// _ "github.com/Kittengarten/KittenCore/draw"    // 牌堆
	_ "github.com/Kittengarten/KittenCore/eekda" // XX 今天吃什么
	// _ "github.com/Kittengarten/KittenCore/essence" // 精华消息
	_ "github.com/Kittengarten/KittenCore/perf"  // 查看 XX
	_ "github.com/Kittengarten/KittenCore/sfacg" // SF 轻小说报更
	_ "github.com/Kittengarten/KittenCore/stack" // 叠猫猫

	// 群管
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/manager"

	// 定时指令触发器
	_ "github.com/FloatTech/zbputils/job"

	// 外部插件
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ahsai"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_false"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aipaint"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aiwife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/alipayvoice"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/b14"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baidu"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base64gua"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baseamasiro"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cangtoushi"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/choose"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chrev"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/coser"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/danbooru"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dish"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drawlots"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dress"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drift_bottle"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/font"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/gif"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hitokoto"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/image_finder"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jiami"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/kfccrazythursday"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolicon"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/midicreate"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moegoe"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu_calendar"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativesetu"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativewife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nbnhhsh"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nsfw"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qzone"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tarot"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tiangou"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tracemoe"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/translation"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wantquotes"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenben"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenxinAI"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/word_count"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ymgal"

	// _ "github.com/Kittengarten/KittenCore/plugin/kokomi"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_reply"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/thesaurus"

	// WebUI，不需要使用可以注释
	webctrl "github.com/FloatTech/zbputils/control/web"
)

var config = kitten.GetMainConfig() // 主配置

func init() {
	// 启用 WebUI
	go webctrl.RunGui(config.WebUI.URL)
}

func main() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); nil != err {
			zap.S().Error(`主函数有 Bug 喵！`, err)
			debug.PrintStack()
		}
	}()
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
