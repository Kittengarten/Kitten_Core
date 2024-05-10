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
	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName = `eekda2`     // 插件名
	todayFile        = `today.yaml` // 保存今天吃什么的文件
	statFile         = `stat.yaml`  // 保存统计数据的文件
	count            = 5            // 每天餐数
	cEEKDA           = `今天吃什么`
	cRegister        = `注册`
	registerSuccess  = `注册成功喵！`
	isNotAdmin       = `不是管理员，无法注册喵！`
	xx               = `XX`
	help             = cRegister + xx + cEEKDA + `	// 在本群注册` + xx + `今天吃什么（管理员可用）
` + xx + cEEKDA + `	// 获取` + xx + `今日食谱
查询被吃次数	// 查询本人被吃次数`
)

var (
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		Brief:             xx + cEEKDA,
		Help:              help,
		PrivateDataFolder: replyServiceName,
	}).ApplySingle(ctxext.DefaultSingle)
	// 读写锁
	mu sync.RWMutex
)

func init() {
	// XX 今天吃什么
	engine.OnSuffix(cEEKDA, zero.OnlyGroup).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.Group)).Handle(todayMeal)

	// 查询被吃次数
	engine.OnFullMatchGroup([]string{`查询被吃次数`, `查看被吃次数`}, zero.OnlyGroup).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.User)).Handle(getStat)
}

// XX 今天吃什么
func todayMeal(ctx *zero.Ctx) {
	cp := getPath(todayFile) // 路径
	var c config
	mu.Lock()
	defer mu.Unlock()
	if err := get(cp, &c); nil != err {
		kitten.SendWithImageFail(ctx, err)
	}
	name := core.MidText(``, cEEKDA, ctx.Event.RawMessage)
	name, needRegister := strings.CutPrefix(name, cRegister)
	ci := slices.IndexFunc(c, func(t today) bool {
		return name == t.ID
	})
	if -1 == ci {
		if needRegister {
			if !zero.AdminPermission(ctx) {
				kitten.SendWithImageFail(ctx, true, isNotAdmin)
				return
			}
			// 注册
			c = append(c, today{
				ID:    name,
				Group: []int64{ctx.Event.GroupID},
			})
			// 写入文件
			if err := set(c, cp); nil != err {
				kitten.SendWithImageFail(ctx, err)
				return
			}
			kitten.SendText(ctx, true, name+registerSuccess)
			return
		}
		kitten.SendWithImageFail(ctx, name+`未在任何群注册喵！`)
		return
	}
	// 成功获取到了角色
	if !slices.Contains(c[ci].Group, ctx.Event.GroupID) {
		// 该角色未在本群注册
		if needRegister {
			if !zero.AdminPermission(ctx) {
				kitten.SendWithImageFail(ctx, true, isNotAdmin)
				return
			}
			// 注册
			c[ci].Group = append(c[ci].Group, ctx.Event.GroupID)
			// 写入文件
			if err := set(c, cp); nil != err {
				kitten.SendWithImageFail(ctx, err)
				return
			}
			kitten.SendText(ctx, true, name+registerSuccess)
			return
		}
		kitten.SendWithImageFail(ctx, name+`未在本群注册喵！`)
		return
	}
	// 该角色已在本群注册
	if needRegister {
		kitten.SendWithImageFail(ctx, name+`已在本群注册，无需重复注册喵！`)
		return
	}
	// 写入上下文
	c[ci].ctx = ctx
	if core.IsSameDate(c[ci].Time, time.Unix(ctx.Event.Time, 0)) {
		// 今天已经生成了，直接播报
		kitten.SendText(ctx, true, &c[ci])
		return
	}
	// 今天没有生成，执行生成
	var list []gjson.Result
	// 获取该角色注册的所有群的群员列表
	for _, v := range c[ci].Group {
		list = append(list, ctx.GetGroupMemberListNoCache(v).Array()...)
	}
	// 只保留昨天一天的群员
	list = slices.DeleteFunc(list, func(v gjson.Result) bool {
		return !core.IsSameDate(time.Unix(v.Get(kitten.LastSentTime).Int(), 0),
			time.Unix(ctx.Event.Time, 0).AddDate(0, 0, -1))
	})
	// 在其中取足够人的下标
	nums, err := core.GenerateRandomNumber(0, len(list), count)
	if err != nil {
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
	if err := set(c, cp); nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	// 播报今天吃什么
	kitten.SendText(ctx, true, &c[ci])
	// 统计
	doStat(ctx, c[ci])
}

// 生成每一餐的内容
func line(ctx *zero.Ctx, u kitten.QQ) string {
	return u.TitleCardOrNickName(ctx) + `	♥	` + u.String()
}

// 加载并反序列化
func get[T *config | *stat](p core.Path, s T) error {
	cd, err := load(p) // 加载
	if nil != err {
		return err
	}
	return unmarshal(cd, s) // 反序列化
}

// 加载
func load(p core.Path) ([]byte, error) {
	if err := core.InitFile(&p, core.Empty); nil != err { // 初始化文件
		return nil, err
	}
	return p.Read() // 读取
}

// 反序列化
func unmarshal[T *config | *stat](d []byte, s T) error {
	if core.Empty == string(d) {
		return nil
	}
	return yaml.Unmarshal(d, s)
}

// 序列化并写入
func set[T config | stat](s T, p core.Path) error {
	b, err := yaml.Marshal(s) // 序列化
	if nil != err {
		return err
	}
	return p.Write(b) // 写入
}

// 获取路径
func getPath(name string) core.Path {
	return core.FilePath(engine.DataFolder(), name)
}
