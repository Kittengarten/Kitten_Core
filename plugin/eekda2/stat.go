package eekda2

import (
	"cmp"
	"maps"
	"slices"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// 比较结构体
type compare struct {
	sum, max, min int
}

// 查询被吃次数
func getStat(ctx *zero.Ctx) {
	mu.RLock()
	defer mu.RUnlock()
	s, err := core.Load[stat](statPath, core.Empty)
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
	}
	if i := slices.IndexFunc(s, func(f food) bool {
		return ctx.Event.UserID == f.ID.Int()
	}); 0 <= i {
		c, err := core.Load[config](todayPath, core.Empty)
		if nil != err {
			kitten.SendWithImageFail(ctx, err)
		}
		for _, t := range c {
			if slices.Contains(t.Group, ctx.Event.GroupID) {
				// 如果当前角色在本群已注册，跳过
				continue
			}
			// 如果当前角色在本群未注册，移除
			maps.DeleteFunc(s[i].Stat, func(k string, v [count]int) bool {
				return k == t.ID
			})
		}
		kitten.SendText(ctx, true, &s[i])
		return
	}
	kitten.DoNotKnow(ctx)
}

// 统计被吃次数
func doStat(ctx *zero.Ctx, td today) {
	s, err := core.Load[stat](statPath, core.Empty)
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
	}
	var ok [count]bool
	// 查询 QQ
	for k, v := range s {
		// 用餐类型
		m := slices.Index(td.Meal[:], v.ID)
		if 0 <= m {
			// 用餐类型有效
			a := s[k].Stat[td.ID]
			a[m]++
			s[k].Stat[td.ID] = a
			ok[m] = true
		}
	}
	// 未查询到的进行写入
	for m, v := range ok {
		if v {
			continue
		}
		var a [count]int
		a[m] = 1
		s = append(s, food{
			ID: td.Meal[m],
			Stat: map[string][count]int{
				td.ID: a,
			},
		})
	}
	// 排序
	s.sort()
	// 写入文件
	if err := core.Save(statPath, s); nil != err {
		kitten.SendWithImageFail(ctx, err)
	}
}

// 统计数据排序
func (s *stat) sort() {
	// 统计数据按总被吃次数排序
	slices.SortStableFunc(*s, func(i, j food) int {
		var (
			ic = i.cmpStat().sum
			jc = j.cmpStat().sum
		)
		if ic < jc {
			return -1
		}
		if ic > jc {
			return 1
		}
		// 如果总数相等，比较集齐五餐的数量
		if c := cmp.Compare(i.cmpStat().min, j.cmpStat().min); 0 != c {
			return c
		}
		// 如果集齐五餐的数量相等，比较单次最高
		return cmp.Compare(i.cmpStat().max, j.cmpStat().max)
	})
}

// 比较
func (f food) cmpStat() (c compare) {
	for _, v := range f.Stat {
		for _, n := range v {
			c.sum += n
			c.max = max(c.max, n)
			c.min = min(c.min, n)
		}
	}
	return
}
