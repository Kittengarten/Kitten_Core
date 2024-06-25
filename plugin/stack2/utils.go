package stack2

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 获取平地摔或特效的概率
func chanceFlat(k meow) float64 {
	return min(float64(mapMeow[抱枕].weight)/float64(k.Weight), 1)
}

// String 实现 fmt.Stringer
func (m meow) String() string {
	if cockroach == globalLocation {
		return fmt.Sprintf(`【%s】	翼展 %.1f cm`, mapMeow[m.getTypeID(globalCtx)].str, itof(m.Weight))
	}
	return fmt.Sprintf(l10nReplacer(globalLocation).Replace(`%s	❤	%d	❤	%.1f kg	%s`),
		m.TitleCardOrNickName(globalCtx), m.Int(), itof(m.Weight), mapMeow[m.getTypeID(globalCtx)].str)
}

// 获取 m 摔坏 n 的概率
func (m meow) chanceFall(n meow) float64 {
	return math.Pow(float64(m.Weight)/float64(m.Weight+n.Weight),
		math.Ln2/(math.Log(math.E+1.0)-1))
}

/*
检查是否因为叠猫猫失败摔下去

m 为上方的猫猫，n 为下方的猫猫

如果没有摔下去则返回 true
*/
func (m meow) checkFall(n meow) bool {
	return m.chanceFall(n) <= rand.Float64()
}

// 获取猫猫类型
func (m meow) getTypeID(ctx *zero.Ctx) meowTypeID {
	for i := range unknown {
		if m.Weight < mapMeow[i].weight {
			if 猫娘少女 == i && m.IsAdult(ctx) {
				continue
			}
			return i
		}
	}
	return unknown
}

// 整数体重转换为浮点（千克数）
func itof(w int) float64 {
	return float64(w) / 10
}

// 浮点体重（千克数）转换为整数
func ftoi(w float64) int {
	return int(10 * w)
}

// 返回服从正态分布 N(0, σ²) 的随机数的绝对值，相当于此分布的右半边
func normal(σ float64) float64 {
	return σ * math.Abs(rand.NormFloat64())
}

// 遍历断言
func rangeAssertion(a []any) []any {
	for k, v := range a {
		switch v := v.(type) {
		case error:
			a[k] = l10nReplacer(globalLocation).Replace(v.Error())
		case fmt.Stringer:
			a[k] = l10nReplacer(globalLocation).Replace(v.String())
		case string:
			a[k] = l10nReplacer(globalLocation).Replace(v)
		}
	}
	return a
}

// 发送本地化文本
func sendText(ctx *zero.Ctx, lf bool, text ...any) message.MessageID {
	return kitten.SendText(ctx, lf, rangeAssertion(text)...)
}

// 发送本地化格式化文本
func sendTextOf(ctx *zero.Ctx, lf bool, format string, a ...any) message.MessageID {
	return kitten.SendTextOf(ctx, lf, l10nReplacer(globalLocation).Replace(format), rangeAssertion(a)...)
}

// 发送带有失败图片的本地化文字消息
func sendWithImageFail(ctx *zero.Ctx, text ...any) message.MessageID {
	return kitten.SendWithImageFail(ctx, rangeAssertion(text)...)
}

// 发送带有自定义图片的本地化文字消息
func sendWithImage(ctx *zero.Ctx, name core.Path, text ...any) message.MessageID {
	return kitten.SendWithImage(ctx, name, rangeAssertion(text)...)
}
