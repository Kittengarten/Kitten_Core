package stack2

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// 叠猫猫排行榜
func (d *data) rank(ctx *zero.Ctx) {
	// 将猫猫按体重顺序排序
	slices.SortFunc(*d, func(i, j meow) int {
		if i.Weight < j.Weight {
			return -1
		}
		if i.Weight > j.Weight {
			return 1
		}
		return cmp.Compare(j.Int(), i.Int())
	})
	var (
		c = len(*d) // 猫猫总数
		r = slices.IndexFunc(*d, func(k meow) bool {
			return ctx.Event.UserID == k.Int()
		}) // 发起查询的猫猫排名
	)
	if 0 == c {
		sendWithImageFail(ctx, `还没有猫猫喵！`)
	}
	if -1 == r {
		sendWithImageFail(ctx, `你没有加入过喵！`)
	}
	var (
		w = make([]int, c, c) // 保存累计重量的切片
		q struct {
			绒布球, 奶猫, 抱枕, 小可爱, 大可爱 int
		} // 数量统计
	)
	for i, m := range *d {
		switch m.getTypeID(ctx) {
		case 绒布球:
			q.绒布球++
		case 奶猫:
			q.奶猫++
		case 抱枕:
			q.抱枕++
		case 小可爱:
			q.小可爱++
		case 大可爱:
			q.大可爱++
		}
		if 0 == i {
			w[i] = m.Weight
			continue
		}
		w[i] = w[i-1] + m.Weight
	}
	var (
		a  = w[c-1] // 猫猫总重量
		wi int      // 重量积分
	)
	for i, v := range w {
		wi += i * v
	}
	s := (*d)[c-10:] // 叠猫猫排行
	sendTextOf(ctx, true, `【叠猫猫排行】
你的当前体重为 %.1f kg
在 %d 只猫猫中排行第 %d 名
所有猫猫当前的总重量为 %.1f kg%s
猫猫体重的基尼系数为 %.3f
猫娘以上：	%d	只
大可爱：　	%d	只
小可爱：　	%d	只
抱枕：　　	%d	只
奶猫：　　	%d	只
绒布球：　	%d	只`,
		itof((*d)[r].Weight),
		c, c-r,
		itof(a),
		func() string {
			if !zero.UserOrGrpAdmin(ctx) {
				return ``
			}
			return fmt.Sprintf(`
————%s`, &s)
		}(),
		1-2*float64(wi)/math.Pow(float64(c), 2)/float64(a),
		c-q.绒布球-q.奶猫-q.抱枕-q.小可爱-q.大可爱,
		q.大可爱,
		q.小可爱,
		q.抱枕,
		q.奶猫,
		q.绒布球,
	)
}
