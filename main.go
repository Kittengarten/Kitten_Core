// KittenCore 的主函数所在包
package main

import (
	// 内置库
	"runtime/debug"

	// KittenCore 的核心库
	"github.com/Kittengarten/KittenCore/internal/protocol"
	"github.com/Kittengarten/KittenCore/kitten"

	// 内部插件
	// _ "github.com/Kittengarten/KittenCore/internal/auth"    // 内置黑名单控制插件
	// _ "github.com/Kittengarten/KittenCore/plugin/draw"    // 牌堆
	_ "github.com/Kittengarten/KittenCore/plugin/eekda2" // XX 今天吃什么
	// _ "github.com/Kittengarten/KittenCore/plugin/essence" // 精华消息
	_ "github.com/Kittengarten/KittenCore/plugin/repeat" // 喵类的本质
	_ "github.com/Kittengarten/KittenCore/plugin/stack2" // 叠猫猫
	_ "github.com/Kittengarten/KittenCore/plugin/track"  // 小说报更
	_ "github.com/Kittengarten/KittenCore/plugin/view"   // 查看 XX

	// 群管
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/manager"

	// 定时指令触发器
	_ "github.com/FloatTech/zbputils/job"

	// 外部插件
	_ "github.com/FloatTech/ZeroBot-Plugin-Playground/plugin/vote"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ahsai"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aifalse"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aiwife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base16384"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base64gua"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baseamasiro"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/choose"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chouxianghua"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chrev"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cpstory"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dailynews"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/danbooru"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dish"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drawlots"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/driftbottle"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/emojimix"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/event"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/font"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/funny"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/gif"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hitokoto"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/inject"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jandan"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolicon"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolimi"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/magicprompt"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/mcfish"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/midicreate"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyucalendar"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qzone"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/realcugan"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/reborn"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shadiao"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shindan"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/sleepmanage"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tarot"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tiangou"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tracemoe"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/translation"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wantquotes"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenxinvilg"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ymgal"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/yujn"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_reply"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/thesaurus"

	// WebUI，不需要使用可以注释
	webctrl "github.com/FloatTech/zbputils/control/web"
)

func init() {
	// 启用 WebUI，不需要使用可以注释
	go webctrl.RunGui(kitten.MainConfig().WebUI.URL)
}

func main() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); nil != err {
			kitten.Error(`主函数有 Bug 喵！`, err, string(debug.Stack()))
		}
	}()
	protocol.RunBot()
}
