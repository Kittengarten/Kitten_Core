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
		ctx   *zero.Ctx        `yaml:"-"`     // 上下文
		Time  time.Time        `yaml:"time"`  // 更新时间
		ID    string           `yaml:"id"`    // 角色名
		Group []int64          `yaml:"group"` // 该角色对应的群号
		Meal  [count]kitten.QQ `yaml:"meal"`  // 今天的每一餐
	}

	// 统计数据集合
	stat []food

	// 食物数据
	food struct {
		ID   kitten.QQ             `yaml:"id"`   // QQ
		Stat map[string][count]int `yaml:"stat"` // 每个角色的个人统计数据
	}
)

// String 实现 fmt.Stringer，播报今天吃什么
func (td *today) String() string {
	return fmt.Sprintf(`【%s今天吃什么】
早餐：　	%s
午餐：　	%s
下午茶：	%s
晚餐：　	%s
夜宵：　	%s`,
		td.ID,
		line(td.ctx, td.Meal[0]),
		line(td.ctx, td.Meal[1]),
		line(td.ctx, td.Meal[2]),
		line(td.ctx, td.Meal[3]),
		line(td.ctx, td.Meal[4]),
	)
}

// String 实现 fmt.Stringer，播报今天吃什么
func (fd *food) String() string {
	var r strings.Builder
	for id, v := range fd.Stat {
		r.WriteString(`【` + id + `】`)
		r.WriteByte('\n')
		r.WriteString(fmt.Sprintf(`早餐：　	%d 次`, v[0]))
		r.WriteByte('\n')
		r.WriteString(fmt.Sprintf(`午餐：　	%d 次`, v[1]))
		r.WriteByte('\n')
		r.WriteString(fmt.Sprintf(`下午茶：	%d 次`, v[2]))
		r.WriteByte('\n')
		r.WriteString(fmt.Sprintf(`晚餐：　	%d 次`, v[3]))
		r.WriteByte('\n')
		r.WriteString(fmt.Sprintf(`夜宵：　	%d 次`, v[4]))
		r.WriteByte('\n')
	}
	return r.String()[:r.Len()-1]
}
