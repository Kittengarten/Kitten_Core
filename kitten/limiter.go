package kitten

import (
	"time"

	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

type limiterType byte // 限速器类型

const (
	GroupNormal limiterType = iota // 群内限速，每 3 分钟 1 次
	GroupFast                      // 群内防刷屏限速，每 12 秒 1 次
	GroupSlow                      // 群内慢限速，每小时 1 次
	User                           // 个人限速，每 12 分钟 1 次
)

var limiter = map[limiterType]func(ctx *zero.Ctx) *rate.Limiter{
	GroupNormal: ctxext.NewLimiterManager(3*time.Minute, 5).LimitByGroup,
	GroupFast:   ctxext.NewLimiterManager(12*time.Second, 5).LimitByGroup,
	GroupSlow:   ctxext.NewLimiterManager(time.Hour, 5).LimitByGroup,
	User:        ctxext.NewLimiterManager(12*time.Minute, 5).LimitByUser,
} // 共通限速器

// GetGetLimiter 获取共通限速器，o 为限速器类型
func GetLimiter(o limiterType) func(ctx *zero.Ctx) *rate.Limiter {
	lmt, ok := limiter[o]
	if ok {
		return lmt
	}
	// 如果获取限速器失败，则返回默认的个人限速器
	Error(`获取限速器失败，请检查限速器类型喵！`)
	return ctxext.LimitByUser
}
