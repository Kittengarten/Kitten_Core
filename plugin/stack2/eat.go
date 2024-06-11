package stack2

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 吃猫猫执行逻辑
func eatExe(ctx *zero.Ctx) {
	if !setGlobalLocation(kitten.GetArgs(ctx)) {
		// 设置全局地区标记位，如当前活动未开放则返回
		kitten.SendWithImageFail(ctx, `当前活动未开放喵！`)
		return
	}
	globalCtx = ctx
	Mu.Lock()
	defer Mu.Unlock()
	d, err := core.Load[data](dataPath, core.Empty)
	if nil != err {
		sendWithImageFail(ctx, `加载叠猫猫数据文件时发生错误喵！`, err)
		return
	}
	d.eat(ctx)
	selfEat(ctx, d, p)
	core.RandomDelay(time.Second)
	selfIn(ctx, d, p)
}

// 吃猫猫
func (d *data) eat(ctx *zero.Ctx) message.MessageID {
	// 初始化自身
	k, err := d.pre(ctx)
	if nil != err {
		return message.MessageID{}
	}
	if 小老虎 > k.getTypeID(ctx) {
		return sendWithImageFail(ctx, `老虎才可以吃猫猫——`)
	}
	// 未在叠猫猫的队列
	dn := d.getNoStack()
	// 执行吃猫猫
	if !d.doEat(ctx, &k) {
		return message.MessageID{}
	}
	// 合并当前未叠猫猫与叠猫猫的队列，将老虎追加入切片中
	*d = slices.Concat(dn, *d, data{k})
	// 存储叠猫猫数据
	if err := core.Save(dataPath, d); nil != err {
		return sendWithImageFail(ctx, `存储叠猫猫数据时发生错误喵！`, err)
	}
	return message.MessageID{}
}

// 执行吃猫猫
func (d *data) doEat(ctx *zero.Ctx, k *meow) bool {
	*d = d.getStack() // 正在叠猫猫的队列
	var (
		dr = slices.Clone(*d) // 叠猫猫队列的克隆
		l  = len(dr)          // 叠猫猫队列高度
	)
	if 0 == l {
		// 如果没有猫猫
		sendWithImageFail(ctx, `猫堆中没有猫猫可以吃——`)
		return false
	}
	if 小老虎 <= (*d)[l-1].getTypeID(ctx) {
		// 老虎以上无法被吃
		sendWithImageFail(ctx, `不可以吃老虎——`)
		return false
	}
	var (
		m    = k
		w, c int // 老虎吃到的体重（0.1kg 数）和猫猫数
	)
	// 从队列的最上部开始遍历（后来居上）
	for i := range *d {
		// 下方的猫猫
		n := &(*d)[l-i-1]
		if !m.checkEat(ctx, *n) {
			// 这只猫猫没有被吃，直接结束遍历
			break
		}
		m = n
		c++
		// 去除被吃的猫猫
		exit(ctx, n, eaten, 0 /* 此参数无效 */)
		// 老虎增加被吃的猫猫的体重
		w += n.Weight
	}
	go setCard(ctx, l-c)
	// 老虎进入休息
	exit(ctx, k, eat, w)
	var r strings.Builder
	if 0 == w {
		r.WriteString(fmt.Sprintf(`吃猫猫失败，杂鱼～杂鱼❤需要休息 %s。`,
			core.ConvertTimeDuration(k.Time.Sub(time.Unix(ctx.Event.Time, 0)))))
		doClear(l, c, k.Weight, k, &r)
		r.WriteRune('🐅')
		sendWithImage(ctx, core.Path(zako), &r)
		return true
	}
	r.WriteString(fmt.Sprintf(`吃猫猫成功，你吃掉了 %d 只猫猫！需要休息 %s。`,
		c, core.ConvertTimeDuration(k.Time.Sub(time.Unix(ctx.Event.Time, 0)))))
	doClear(l, c, k.Weight-w, k, &r)
	r.WriteRune('🐯')
	for range c {
		r.WriteRune('😿')
	}
	e := dr[l-c:]
	sendText(ctx, true, &r, &e)
	return true
}

// 检查是否成功吃掉，m 为老虎，n 为猫猫
func (m meow) checkEat(ctx *zero.Ctx, n meow) bool {
	if 小老虎 <= n.getTypeID(ctx) {
		// 老虎不能被吃
		return false
	}
	return m.chanceFall(n) > rand.Float64()
}
