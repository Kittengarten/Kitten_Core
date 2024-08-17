package stack2

import (
	"errors"
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// 评估叠猫猫，返回叠入的权重作为概率
func (d *data) evaluate(ctx *zero.Ctx) float64 {
	var (
		dr     = slices.Clone(*d) // 克隆切片，防止对后续调用造成影响
		k, err = dr.pre(ctx)      // 初始化自身
	)
	if nil != err {
		// 如果不能活动，什么也不做
		return 0
	}
	var (
		s = dr.getStack() // 获取叠猫猫队列
		l = len(s)        // 叠猫猫队列长度
	)
	if 0 == l {
		// 如果是空队列，直接叠入，尝试平地摔
		return 1
	}
	// 如果是非空队列
	if float64(l) <= float64(k.Weight)*chanceFlat(k)*chanceClear(ctx, s, k)*(math.E-1) {
		// 如果清空特效导致体重增加的期望不少于当前的猫堆高度，直接叠入
		return 1
	}
	var (
		sn = append(s, k)          // 用于压坏判定的队列
		cp = sn.chancePressed(ctx) // 压坏概率
		gp = func(m meow) float64 {
			if 抱枕 >= k.getTypeID(ctx) || 幼年猫娘 <= m.getTypeID(ctx) {
				// 抱枕及以下的猫猫不会导致猫猫摔下去，直接在猫娘身上叠猫猫不会摔下去
				return 0
			}
			return k.chanceFall(m)
		}(s[l-1]) // 不压坏的情况下，摔下去的概率
		cf = (1 - cp) * gp // 摔下概率

	)
	if 0.5 <= cf {
		// 摔下概率达到 50% 以上，不叠入
		return 0
	}
	if mapMeow[幼年猫娘].weight <= s[0].Weight && 0 < cp {
		// 如果底座是猫娘萝莉以上，只要可能压坏，就不叠入
		return 0
	}
	if 0.5 <= cp {
		// 压猫猫！
		// 如果压坏概率达到 50% 以上，只要队列中没有猫娘萝莉以上，就按照（压坏概率 - 摔下概率）× 自身体重与平均体重 e 倍的比值叠入
		for _, k := range s {
			if mapMeow[幼年猫娘].weight <= k.Weight {
				// 有猫娘萝莉以上，快跑！
				return 0
			}
		}
		return (cp - cf) * float64(k.Weight) / (math.E * float64(s.totalWeight()) / float64(l))
	}
	// 傍大猫
	// 底座越重且层数越低，越应该叠入
	return (1.0-float64(k.Weight)/float64(s[0].Weight))/float64(l) - cf
}

// 自动加入
func selfIn(ctx *zero.Ctx, d data) bool {
	ctx.Event.UserID = ctx.Event.SelfID
	if d.evaluate(ctx) <= rand.Float64() {
		// 以评估的概率，触发喵喵使用 /叠猫猫 加入
		return false
	}
	core.RandomDelay(time.Second)
	ctx.Send(botConfig.CommandPrefix + cStack + cMeow + ` ` + cIn)
	core.RandomDelay(time.Second)
	d.in(ctx)
	return true
}

// 自动分析
func selfAnalysis(ctx *zero.Ctx, d data) {
	ctx.Event.UserID = ctx.Event.SelfID
	if d.evaluate(ctx) <= rand.Float64() {
		// 以评估的概率，触发喵喵使用 /叠猫猫 分析
		return
	}
	core.RandomDelay(time.Second)
	ctx.Send(botConfig.CommandPrefix + cStack + cMeow + ` ` + cAnalysis)
	core.RandomDelay(time.Second)
	d.analysis(ctx)
}

// 自动排行
func selfRank(ctx *zero.Ctx, d data) {
	ctx.Event.UserID = ctx.Event.SelfID
	if 0.1 < rand.Float64() {
		// 以 0.1 的概率，触发喵喵使用 /叠猫猫 排行
		return
	}
	core.RandomDelay(time.Second)
	ctx.Send(botConfig.CommandPrefix + cStack + cMeow + ` ` + cRank)
	core.RandomDelay(time.Second)
	d.rank(ctx)
}

// 评估吃猫猫，返回吃的权重作为概率
func (d *data) evaluateEat(ctx *zero.Ctx) float64 {
	var (
		dr     = slices.Clone(*d) // 克隆切片，防止对后续调用造成影响
		k, err = dr.pre(ctx)      // 初始化自身
	)
	if nil != err {
		// 如果不能活动，什么也不做
		return 0
	}
	var (
		s = dr.getStack() // 获取叠猫猫队列
		l = len(s)        // 叠猫猫队列长度
	)
	if 0 == l {
		// 如果是空队列，什么也不做
		return 0
	}
	// 如果是非空队列，吃猫猫的概率为期望占小老虎体重的比例 - 0.5
	return float64(s[l-1].Weight)*k.chanceFall(s[l-1])/itof(mapMeow[猫娘少女].weight) - 0.5
}

// 自动吃猫猫
func selfEat(ctx *zero.Ctx, d data) bool {
	ctx.Event.UserID = ctx.Event.SelfID
	if d.evaluateEat(ctx) < rand.Float64() {
		// 以评估的概率，触发喵喵使用 /吃猫猫
		return false
	}
	core.RandomDelay(time.Second)
	ctx.Send(botConfig.CommandPrefix + cEat + cMeow)
	core.RandomDelay(time.Second)
	d.eat(ctx)
	return true
}

// 评估加速，返回加速的权重作为概率
func (d *data) evaluateOC(ctx *zero.Ctx) float64 {
	var (
		dr     = slices.Clone(*d)  // 克隆切片，防止对后续调用造成影响
		_, err = dr.pre(ctx)       // 初始化自身
		nre    = needRest(0, 0, 0) // 默认错误：需要休息
	)
	if !errors.As(err, &nre) {
		// 如果当前不在休息，不需要加速，什么也不做
		return 0
	}
	nre = err.(*needRestErr) // 需要休息
	if 大老虎 > (*d)[nre.i].getTypeID(ctx) {
		// 如果不是大老虎，不能加速，什么也不做
		return 0
	}
	var (
		omrt  = time.Hour * time.Duration(stackConfig.OCMinRestHours)           // 最小休息时间
		hours = int(math.RoundToEven(float64(nre.t-omrt) / float64(time.Hour))) // 加速的小时数
	)
	if 0 >= hours {
		// 如果加速的小时数不大于 0，则不能加速，什么也不做
		return 0
	}
	if (*d)[nre.i].Weight-hours < 1 {
		// 如果体重不足，则不能加速，什么也不做
		return 0
	}
	// 加速的权重为 (ln(当前体重（0.1 kg 数） ÷ 加速的小时数) - e) ÷ e^e
	return math.Log(float64((*d)[nre.i].Weight)/float64(hours)-math.E) / math.Pow(math.E, math.E)
}

// 自动加速
func selfOC(ctx *zero.Ctx, d data) bool {
	ctx.Event.UserID = ctx.Event.SelfID
	if d.evaluateOC(ctx) < rand.Float64() {
		// 以评估的概率，触发喵喵使用 /叠猫猫 锻炼
		return false
	}
	core.RandomDelay(time.Second)
	ctx.Send(botConfig.CommandPrefix + cStack + cMeow + ` ` + cOCCat)
	core.RandomDelay(time.Second)
	d.oc(ctx)
	return true
}
