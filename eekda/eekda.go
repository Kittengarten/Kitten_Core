package eekda

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName = `eekda`      // 插件名
	todayFile        = `today.yaml` // 保存今天吃什么的文件
	statFile         = `stat.yaml`  // 保存统计数据的文件
	ee               = `翼翼`         // 默认名字
	cEEKDA           = `今天吃什么`
)

var namePath = kitten.FilePath(replyServiceName, `name.txt`) // 保存名字的文件

func init() {
	var (
		name  = namePath.GetString(ee)
		brief = fmt.Sprint(name, cEEKDA)
		help  = fmt.Sprintf(`%s%s // 获取%s今日食谱
查询被吃次数 // 查询本人被吃次数`,
			name, cEEKDA, name)
		// 注册插件
		engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  true,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: replyServiceName,
		}).ApplySingle(ctxext.DefaultSingle)
	)

	engine.OnFullMatch(fmt.Sprint(name, cEEKDA), zero.OnlyGroup).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Hour, 1).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		tf := kitten.FilePath(engine.DataFolder(), todayFile) // 保存今天吃什么的文件路径
		kitten.InitFile(&tf, kitten.Empty)                    // 初始化文件
		t, err := tf.Read()
		if nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		var td today
		if err := yaml.Unmarshal(t, &td); nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		if kitten.IsSameDate(time.Now(), td.Time) {
			report(ctx, td, name)
			return
		}
		// 生成今天吃什么
		list := ctx.GetThisGroupMemberListNoCache().Array()
		slices.SortStableFunc(list, func(i, j gjson.Result) int {
			return cmp.Compare(i.Get(`last_sent_time`).Int(), j.Get(`last_sent_time`).Int())
		})
		list = list[max(0, len(list)-50):]
		nums := kitten.GenerateRandomNumber(0, len(list), 5)
		if 5 > len(nums) {
			ctx.Send(`没有足够的食物喵！`)
			return
		}
		td = today{
			Time:      time.Now(),
			Breakfast: list[nums[0]].Get(`user_id`).Int(),
			Lunch:     list[nums[1]].Get(`user_id`).Int(),
			LowTea:    list[nums[2]].Get(`user_id`).Int(),
			Dinner:    list[nums[3]].Get(`user_id`).Int(),
			Supper:    list[nums[4]].Get(`user_id`).Int(),
		}
		if t, err = yaml.Marshal(td); nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		if err = tf.Write(t); nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		report(ctx, td, name)
		// 存储饮食统计数据
		sf := kitten.FilePath(engine.DataFolder(), statFile)
		s, err := sf.Read()
		if nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		var sd stat
		if err := yaml.Unmarshal(s, &sd); nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		sm := make(map[int64]food, len(sd)) // QQ:猫猫集合
		// 加载并修改数据
		for k := range sd {
			sm[sd[k].ID] = sd[k]
		}
		bf, bfok := sm[td.Breakfast]
		if bfok {
			bf.Breakfast++
			sm[td.Breakfast] = bf
		} else {
			sd = append(sd, newFoodToday(ctx, td, breakfast))
		}
		l, lok := sm[td.Lunch]
		if lok {
			l.Lunch++
			sm[td.Lunch] = l
		} else {
			sd = append(sd, newFoodToday(ctx, td, lunch))
		}
		lt, ltok := sm[td.LowTea]
		if ltok {
			lt.LowTea++
			sm[td.LowTea] = lt
		} else {
			sd = append(sd, newFoodToday(ctx, td, lowtea))
		}
		d, dok := sm[td.Dinner]
		if dok {
			d.Dinner++
			sm[td.Dinner] = d
		} else {
			sd = append(sd, newFoodToday(ctx, td, dinner))
		}
		sp, spok := sm[td.Supper]
		if spok {
			sp.Supper++
			sm[td.Supper] = sp
		} else {
			sd = append(sd, newFoodToday(ctx, td, supper))
		}
		// 回写修改的数据
		for k := range sd {
			sd[k] = sm[sd[k].ID]
		}
		// 统计数据按总被吃次数排序
		slices.SortStableFunc(sd, func(i, j food) int {
			ic := i.Breakfast + i.Lunch + i.LowTea + i.Dinner + i.Supper
			jc := j.Breakfast + j.Lunch + j.LowTea + j.Dinner + j.Supper
			if ic < jc {
				return -1
			}
			if ic > jc {
				return 1
			}
			// 如果总数相等，比较集齐五餐的数量
			if c := cmp.Compare(min(i.Breakfast, i.Lunch, i.LowTea, i.Dinner, i.Supper), min(j.Breakfast, j.Lunch, j.LowTea, j.Dinner, j.Supper)); 0 != c {
				return c
			}
			// 如果集齐五餐的数量相等，比较单次最高
			return cmp.Compare(max(i.Breakfast, i.Lunch, i.LowTea, i.Dinner, i.Supper), max(j.Breakfast, j.Lunch, j.LowTea, j.Dinner, j.Supper))
		})
		s, err = yaml.Marshal(sd)
		sf.Write(s)
		if nil != err {
			zap.S().Error(`写入饮食统计数据发生错误：`, err)
		}
	})

	engine.OnFullMatchGroup([]string{`查询被吃次数`, `查看被吃次数`}, zero.OnlyGroup).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Hour, 2).LimitByUser).Handle(func(ctx *zero.Ctx) {
		sf := kitten.FilePath(engine.DataFolder(), statFile)
		kitten.InitFile(&sf, kitten.Empty)
		s, err := sf.Read()
		if nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		var sd stat
		if err := yaml.Unmarshal(s, &sd); nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		if nil != yaml.Unmarshal(s, &sd) {
			zap.S().Error(`饮食统计数据损坏了喵！`)
			kitten.DoNotKnow(ctx)
			return
		}
		for i := range sd {
			if ctx.Event.UserID == sd[i].ID {
				kitten.SendText(ctx, true, fmt.Sprintf(`【%s的被吃次数】
早餐：%d 次
午餐：%d 次
下午茶：%d 次
晚餐：%d 次
夜宵：%d 次`,
					getLine(ctx, kitten.QQ(sd[i].ID)),
					sd[i].Breakfast,
					sd[i].Lunch,
					sd[i].LowTea,
					sd[i].Dinner,
					sd[i].Supper))
				return
			}
		}
		kitten.DoNotKnow(ctx)
	})
}

// 获取条目，u 为 QQ
func getLine(ctx *zero.Ctx, u kitten.QQ) string {
	return fmt.Sprintf(`%s（%d）`, u.GetTitleCardOrNickName(ctx), u)
}

// 播报今天吃什么，td 为今日数据
func report(ctx *zero.Ctx, td today, name string) {
	ctx.Send(fmt.Sprintf(`【%s今天吃什么】
早餐:　%s
午餐:　%s
下午茶:%s
晚餐:　%s
夜宵:　%s`,
		name,
		newFoodToday(ctx, td, breakfast).Name,
		newFoodToday(ctx, td, lunch).Name,
		newFoodToday(ctx, td, lowtea).Name,
		newFoodToday(ctx, td, dinner).Name,
		newFoodToday(ctx, td, supper).Name))
}
