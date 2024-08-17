package eekda2

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/tidwall/gjson"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName  = `eekda2`     // 插件名
	todayFile         = `today.yaml` // 保存今天吃什么的文件
	statFile          = `stat.yaml`  // 保存统计数据的文件
	count             = 5            // 每天餐数
	cEEKDA            = `今天吃什么`
	cRegister         = `注册`
	cUnregister       = `注销`
	registerSuccess   = cRegister + `成功喵！`
	unregisterSuccess = cUnregister + `成功喵！`
	isNotAdmin        = `不是管理员，无法操作喵！`
	xx                = `XX`
	thisGroup         = ` // 在本群`
	help              = cRegister + xx + cEEKDA + thisGroup + cRegister + xx + `今天吃什么（管理员可用）
` + cUnregister + xx + cEEKDA + thisGroup + cUnregister + xx + `今天吃什么（管理员可用）
` + xx + cEEKDA + ` // 获取` + xx + `今日食谱
查询被吃次数 // 查询本人被吃次数`
)

var (
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		Brief:             xx + cEEKDA,
		Help:              help,
		PrivateDataFolder: replyServiceName,
	}).ApplySingle(ctxext.DefaultSingle)
	// 今日文件路径
	todayPath = core.FilePath(engine.DataFolder(), todayFile)
	// 统计文件路径
	statPath = core.FilePath(engine.DataFolder(), statFile)
	// 读写锁
	mu sync.RWMutex
)

func init() {
	// XX 今天吃什么
	engine.OnSuffix(cEEKDA, zero.OnlyGroup).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.GroupNormal)).Handle(todayMeal)

	// 查询被吃次数
	engine.OnFullMatchGroup([]string{`查询被吃次数`, `查看被吃次数`}, zero.OnlyGroup).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.User)).Handle(getStat)
}

// XX 今天吃什么
func todayMeal(ctx *zero.Ctx) {
	mu.Lock()
	defer mu.Unlock()
	c, err := core.Load[config](todayPath, core.Empty)
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
	}
	name := core.MidText(``, cEEKDA, core.CleanAll(ctx.Event.RawMessage, false))
	name, needRegister := strings.CutPrefix(name, cRegister)
	name, needUnegister := strings.CutPrefix(name, cUnregister)
	if `` == name {
		// 角色名为空
		kitten.SendWithImageFail(ctx, `角色名为空喵！`)
		return
	}
	if needRegister && needUnegister {
		// 指令冲突
		kitten.SendWithImageFail(ctx, `指令冲突喵！`)
		return
	}
	ci := slices.IndexFunc(c, func(t today) bool {
		return name == t.ID
	})
	// 角色是否存在
	if -1 == ci {
		// 该角色不存在
		if !needRegister {
			// 不执行注册指令
			kitten.SendWithImageFail(ctx, name+`未在任何群注册喵！`)
			return
		}
		// 执行注册指令
		if !zero.AdminPermission(ctx) {
			// 没有权限
			kitten.SendWithImageFail(ctx, isNotAdmin)
			return
		}
		// 注册
		c = append(c, today{
			ID:    name,
			Group: []int64{ctx.Event.GroupID},
		})
		// 写入文件
		if err := core.Save(todayPath, c); nil != err {
			kitten.SendWithImageFail(ctx, err)
			return
		}
		kitten.SendText(ctx, true, name+registerSuccess)
		return
	}
	// 该角色存在
	if !slices.Contains(c[ci].Group, ctx.Event.GroupID) {
		// 该角色未在本群注册
		if !needRegister {
			// 不执行注册指令
			kitten.SendWithImageFail(ctx, name+`未在本群注册喵！`)
			return
		}
		// 执行注册指令
		if !zero.AdminPermission(ctx) {
			// 没有权限
			kitten.SendWithImageFail(ctx, isNotAdmin)
			return
		}
		// 注册
		c[ci].Group = append(c[ci].Group, ctx.Event.GroupID)
		// 写入文件
		if err := core.Save(todayPath, c); nil != err {
			kitten.SendWithImageFail(ctx, err)
			return
		}
		kitten.SendText(ctx, true, name+registerSuccess)
		return
	}
	// 该角色已在本群注册
	switch {
	case needRegister:
		// 执行注册指令
		kitten.SendWithImageFail(ctx, name+`已在本群注册，无需重复注册喵！`)
	case needUnegister:
		// 执行注销指令
		if !zero.AdminPermission(ctx) {
			// 没有权限
			kitten.SendWithImageFail(ctx, isNotAdmin)
			return
		}
		// 注销
		c[ci].Group = slices.DeleteFunc(c[ci].Group, func(groupID int64) bool {
			return groupID == ctx.Event.GroupID
		})
		if 0 == len(c[ci].Group) {
			// 如果该角色已经在所有群注销，删除该角色
			c = slices.DeleteFunc(c, func(t today) bool {
				return name == t.ID
			})
		}
		// 写入文件
		if err := core.Save(todayPath, c); nil != err {
			kitten.SendWithImageFail(ctx, err)
			return
		}
		kitten.SendText(ctx, true, name+unregisterSuccess)
	default:
		// 执行通常指令，写入上下文
		c[ci].ctx = ctx
		if core.IsSameDate(c[ci].Time, time.Unix(ctx.Event.Time, 0)) {
			// 今天已经生成了，直接播报
			kitten.SendText(ctx, true, &c[ci])
			return
		}
		// 今天没有生成，执行生成
		list := make([]gjson.Result, 0, 128)
		// 获取该角色注册的所有群的群员列表
		for _, v := range c[ci].Group {
			list = append(list, kitten.MemberList(ctx, v).List...)
		}
		// 只保留昨天一天的群员
		list = slices.DeleteFunc(list, func(v gjson.Result) bool {
			return !core.IsSameDate(time.Unix(v.Get(kitten.LastSentTime).Int(), 0),
				time.Unix(ctx.Event.Time, 0).AddDate(0, 0, -1))
		})
		// 在其中取足够人的下标
		nums, err := core.GenerateRandomNumber(0, len(list), count)
		if nil != err {
			kitten.SendWithImageFail(ctx, fmt.Errorf(`没有足够的食物喵！%w`, err))
			return
		}
		// 传入足够人的 QQ
		for i, v := range nums {
			c[ci].Meal[i] = kitten.QQ(list[v].Get(kitten.UserID).Int())
		}
		// 写入时间
		c[ci].Time = time.Unix(ctx.Event.Time, 0)
		// 写入文件
		if err := core.Save(todayPath, c); nil != err {
			kitten.SendWithImageFail(ctx, err)
			return
		}
		// 播报今天吃什么
		kitten.SendText(ctx, true, &c[ci])
		// 统计
		doStat(ctx, c[ci])
	}
}

// 生成每一餐的内容
func line(ctx *zero.Ctx, u kitten.QQ) string {
	return u.TitleCardOrNickName(ctx) + `	❤	` + u.String()
}
