package kitten

import (
	"time"

	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

type limiterType byte // 限速器类型

const (
	Group     limiterType = iota // 群内限速，每 15 分钟 5 次
	GroupFast                    // 群内防刷屏限速，每分钟 5 次
	GroupSlow                    // 群内慢限速，每小时 1 次
	User                         // 个人限速，每小时 5 次
)

var limiter = map[limiterType]func(ctx *zero.Ctx) *rate.Limiter{
	Group:     ctxext.NewLimiterManager(15*time.Minute, 5).LimitByGroup,
	GroupFast: ctxext.NewLimiterManager(time.Minute, 5).LimitByGroup,
	GroupSlow: ctxext.NewLimiterManager(time.Hour, 1).LimitByGroup,
	User:      ctxext.NewLimiterManager(time.Hour, 5).LimitByUser,
} // 限速器

// GetGetLimiter 获取限速器，o 为限速对象
func GetLimiter(o limiterType) func(ctx *zero.Ctx) *rate.Limiter {
	lmt, ok := limiter[o]
	if ok {
		return lmt
	}
	// 如果获取限速器失败，则返回默认的个人限速器
	return ctxext.LimitByUser
}
