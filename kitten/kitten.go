// Package kitten 包含了 KittenCore 以及各插件的核心依赖结构体、方法和函数
package kitten

import (
	"fmt"
	"time"

	"github.com/FloatTech/zbputils/ctxext"
	"github.com/Kittengarten/KittenCore/kitten/core"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"

	"gopkg.in/yaml.v3"
)

const (
	User       byte = iota // 个人限速
	Group                  // 群内限速
	GroupFast              // 群内防刷屏限速
	configFile = `config.yaml`
)

var (
	botConfig config    // 来自 Bot 的配置文件
	imagePath core.Path // 图片路径
	Limiter   limiter   // 限速器
	Weight    int       // 自身叠猫猫体重（0.1kg 数）
)

func init() {
	// 配置文件加载
	d, err := core.Path(configFile).Read()
	if nil != err {
		fmt.Println(err, `请配置 `+configFile+` 后重新启动喵！`)
	}
	if err := yaml.Unmarshal(d, &botConfig); nil != err {
		fmt.Println(err, `请正确配置 `+configFile+` 后重新启动喵！`)
	}
	if 0 == len(botConfig.SuperUsers) {
		fmt.Println(`请在 ` + configFile + ` 中配置 superusers 喵！`)
	}
	// 图片路径
	imagePath = core.FilePath(botConfig.Path, `image`)
	// 定义限速器
	Limiter.l = map[byte]func(ctx *zero.Ctx) *rate.Limiter{
		GroupFast: ctxext.NewLimiterManager(time.Minute, 5).LimitByGroup,
		Group:     ctxext.NewLimiterManager(15*time.Minute, 5).LimitByGroup,
		User:      ctxext.NewLimiterManager(time.Hour, 5).LimitByUser,
	}
}

// MainConfig 获取主配置
func MainConfig() config {
	return botConfig
}

// Get 获取限速器，o 为限速对象
func (l limiter) Get(o byte) func(ctx *zero.Ctx) *rate.Limiter {
	lmt, ok := l.l[o]
	if ok {
		return lmt
	}
	// 如果获取限速器失败，则返回默认的个人限速器
	return ctxext.LimitByUser
}
