package eekda2

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	breakfast mealType = iota // 早餐
	lunch                     // 午餐
	lowtea                    // 下午茶
	dinner                    // 晚餐
	supper                    // 夜宵
)

type (
	// 用餐类型
	mealType byte

	// 配置文件
	config []today

	// 今天吃什么
	today struct {
		ctx   *zero.Ctx        `yaml:"-"` // 上下文
		Time  time.Time        // 更新时间
		ID    string           // 角色名
		Group []int64          // 该角色对应的群号
		Meal  [count]kitten.QQ // 今天的每一餐
	}

	// 统计数据集合
	stat []food

	// 食物数据
	food struct {
		ID   kitten.QQ             // QQ
		Stat map[string][count]int // 每个角色的个人统计数据
	}
)

// String 实现 fmt.Stringer，播报今天吃什么
func (td *today) String() string {
	return `【` + td.ID + `今天吃什么】
早餐：　	` + line(td.ctx, td.Meal[breakfast]) + `
午餐：　	` + line(td.ctx, td.Meal[lunch]) + `
下午茶：	` + line(td.ctx, td.Meal[lowtea]) + `
晚餐：　	` + line(td.ctx, td.Meal[dinner]) + `
夜宵：　	` + line(td.ctx, td.Meal[supper])
}

// String 实现 fmt.Stringer，播报今天吃什么
func (fd *food) String() string {
	var (
		r  strings.Builder
		lf bool
	)
	for id, v := range fd.Stat {
		if lf {
			r.WriteByte('\n')
		} else {
			lf = true
		}
		r.WriteString(`【` + id + "】\n")
		r.WriteString(fmt.Sprintf("早餐：　	%d 次\n", v[breakfast]))
		r.WriteString(fmt.Sprintf("午餐：　	%d 次\n", v[lunch]))
		r.WriteString(fmt.Sprintf("下午茶：	%d 次\n", v[lowtea]))
		r.WriteString(fmt.Sprintf("晚餐：　	%d 次\n", v[dinner]))
		r.WriteString(fmt.Sprintf(`夜宵：　	%d 次`, v[supper]))
	}
	return r.String()
}
