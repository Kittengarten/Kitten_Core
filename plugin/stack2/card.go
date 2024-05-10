package stack2

import (
	"fmt"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var active sync.Map // 各群的上次活跃时间

func setCard(ctx *zero.Ctx, h int) {
	if 0 >= ctx.Event.GroupID {
		return
	}
	// 保存本群的活跃时间
	active.Store(ctx.Event.GroupID, time.Unix(ctx.Event.Time, 0))
	// 如果群距上次活跃时间大于一天，则删除
	active.Range(func(g any, t any) bool {
		if core.HoursPerDay*time.Hour < time.Since(t.(time.Time)) {
			ctx.SetGroupCard(g.(int64), sid.Int(), card(ctx, -1))
			active.Delete(g)
			return true
		}
		ctx.SetGroupCard(g.(int64), sid.Int(), card(ctx, h))
		return true
	})
}

func card(ctx *zero.Ctx, h int) string {
	if 0 < h {
		return fmt.Sprintf(`%s（%d岁）（猫堆高度：%d）`, botConfig.NickName[0], sid.Age(ctx), h)
	}
	if 0 == h {
		return fmt.Sprintf(`%s（%d岁）（猫堆已清空）`, botConfig.NickName[0], sid.Age(ctx))
	}
	return fmt.Sprintf(`%s（%d岁）（内置冷却，禁止调戏）`, botConfig.NickName[0], sid.Age(ctx))
}
