// Package view 查看服务器运行状况
package view

import (
	"math/rand/v2"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	replyServiceName   = `perf` // 插件名
	brief              = `查看运行状况`
	filePath           = `file.txt` // 保存微星小飞机温度配置文件路径的文件，非 Windows 系统或不使用可以忽略
	cView              = `查看`
	defaultTemperature = `45` // 默认温度
)

var (
	// 默认昵称
	nickname = kitten.MainConfig().NickName[0]
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            brief,
		Help: func() string {
			var s strings.Builder // 字符串构建器
			s.Grow(32 * len(kitten.MainConfig().NickName))
			for _, n := range kitten.MainConfig().NickName {
				s.WriteString(kitten.MainConfig().CommandPrefix + cView + n + ` // 可获取服务器运行状况`)
				s.WriteByte('\n')
			}
			s.WriteString(`戳一戳` + nickname + ` // 可得到响应`)
			return s.String()
		}(),
	}).ApplySingle(ctxext.DefaultSingle)
	// 戳一戳限速
	pokeLimiterManager = ctxext.NewLimiterManager(5*time.Minute, 9)
)

func init() {
	if `linux` == runtime.GOOS {
		// signd 守护协程
		go signd()
	}

	// 查看功能
	engine.OnCommand(cView).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.User)).
		Limit(kitten.GetLimiter(kitten.Group)).
		Handle(view)

	// 支付宝到账语音
	engine.OnPrefix(`支付宝到账`).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.User)).
		Limit(kitten.GetLimiter(kitten.Group)).
		Handle(sendAlipayVoice)

	// Ping 功能
	engine.OnCommandGroup([]string{`Ping`, `ping`}, zero.SuperUserPermission).
		SetBlock(true).Limit(ctxext.LimitByGroup).Handle(ping)

	// 戳一戳
	engine.On(`notice/notify/poke`, zero.OnlyToMe).SetBlock(true).Handle(poke)

	// 通过 CQ 码、链接等让 Bot 发送图片，为防止滥用，仅管理员可用
	zero.OnCommand(`图片`, zero.AdminPermission).SetBlock(true).Handle(sendImage)

	// 通过 CQ 码、链接、图片等让 Bot 扫描二维码，为防止滥用，仅管理员可用
	zero.OnCommandGroup([]string{`扫码`, `扫描`}, zero.AdminPermission).SetBlock(true).Handle(scan)
}

// 查看
func view(ctx *zero.Ctx) {
	switch func() (who string) {
		who = core.CleanAll(kitten.GetArgs(ctx), false)
		for _, n := range kitten.MainConfig().NickName {
			if who == n {
				who = nickname
			}
		}
		return
	}() {
	case nickname:
		img, err := core.FilePath(kitten.MainConfig().Path, replyServiceName, `image`).
			Image(core.Path(strconv.Itoa(
				getPerf(cpuPercent(), percent(getMem()), defaultTemperature)) + `.png`))
		if nil != err {
			kitten.SendWithImageFail(ctx, err)
			return
		}
		kitten.SendMessage(ctx, true, img, message.Text(viewString()))
	case `鸡汤`:
		send(ctx, jiTang, false)
	case `情话`:
		send(ctx, qingHua, false)
	case `疯狂星期四`:
		if time.Now().Weekday() != time.Thursday {
			// 如果不是星期四，则不发送
			kitten.SendWithImageFail(ctx, `今天不是星期四喵！`)
			return
		}
		send(ctx, kfc, false)
	case `一言`:
		sendYiYan(ctx)
	case `waifu`, `老婆`, `随机老婆`:
		sendWaifu(ctx)
	default:
		kitten.DoNotKnow(ctx)
	}
}

// 戳一戳
func poke(ctx *zero.Ctx) {
	switch limiter := pokeLimiterManager.LimitByGroup(ctx); {
	case limiter.AcquireN(5):
		// 5 分钟共 9 块命令牌 一次消耗 5 块命令牌
		ctx.Send(message.Poke(ctx.Event.UserID))
	case limiter.AcquireN(3):
		// 5 分钟共 9 块命令牌 一次消耗 3 块命令牌
		kitten.SendWithImageFail(ctx, `请不要拍`+nickname+` >_<`)
	case limiter.Acquire():
		// 5 分钟共 9 块命令牌 一次消耗 1 块命令牌
		kitten.SendWithImageFailOf(ctx, "喂(#`O′) 拍%s干嘛！\n（好感 - %d）", nickname, 1+rand.IntN(100))
		// 频繁触发，不回复
	}
}
