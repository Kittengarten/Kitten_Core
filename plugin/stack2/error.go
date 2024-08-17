package stack2

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type (
	// 已经加入叠猫猫
	alreadyJoinedErr struct{}

	// 需要休息
	needRestErr struct {
		t time.Duration // 剩余的休息时间
		w int           // 叠入猫猫的体重
		i int           // 叠入猫猫的下标
	}

	// 叠猫猫失败
	stackErr struct {
		ctx             *zero.Ctx // 上下文
		k               *meow     // 失败的猫猫
		l               int       // 叠猫猫队列高度
		n               int       // 造成别的猫猫退出的数量
		r               result    // 退出原因
		strings.Builder           // 错误内容
	}
)

// Error 实现 error
func (*alreadyJoinedErr) Error() string {
	return `已经加入叠猫猫了喵！`
}

// *alreadyJoinedErr 的构造函数，已经加入叠猫猫
func alreadyJoined() *alreadyJoinedErr {
	return &alreadyJoinedErr{}
}

// Error 实现 error
func (e *needRestErr) Error() string {
	return fmt.Sprintf(`还需要休息 %s才能活动喵！
你的当前体重为 %.1f kg。`,
		core.ConvertTimeDuration(e.t),
		itof(e.w))
}

// *needRest 的构造函数，需要休息
func needRest(t time.Duration, w, i int) *needRestErr {
	return &needRestErr{
		t: t,
		w: w,
		i: i,
	}
}

// Error 实现 error
func (e *stackErr) Error() string {
	if 0 != e.Len() {
		return e.String()
	}
	w := e.k.Weight // 叠猫猫前的体重
	e.Grow(128)
	e.WriteString(`叠猫猫失败，杂鱼～杂鱼❤`)
	switch e.r {
	case flat:
		// 如果平地摔
		exit(e.ctx, e.k, e.r, e.n) // 让失败的猫猫退出
		e.WriteString(fmt.Sprintf(`你平地摔了喵！需要休息 %s。
你的体重由 %.1f kg 变为 %.1f kg。`,
			core.ConvertTimeDuration(e.k.Time.Sub(time.Unix(e.ctx.Event.Time, 0))),
			itof(w), itof(e.k.Weight)))
	case press:
		// 压坏了别的猫猫
		exit(e.ctx, e.k, e.r, e.n) // 让失败的猫猫退出
		e.WriteString(fmt.Sprintf(`有 %d 只猫猫被压坏了喵！需要休息一段时间。`, e.n))
		doClear(e.l, e.n, w, e.k, &e.Builder)
		for range e.n {
			e.WriteRune('🙀')
		}
	case fall:
		// 摔坏了别的猫猫
		exit(e.ctx, e.k, e.r, e.l) // 让失败的猫猫退出
		e.WriteString(fmt.Sprintf(`上面 %d 只猫猫摔下去了喵！需要休息一段时间。`, e.n))
		doClear(e.l, e.n, w, e.k, &e.Builder)
		for range e.n {
			e.WriteRune('😿')
		}
	default:
		e.WriteString(`未知错误喵！`)
	}
	return e.String()
}

/*
*stackErr 的构造函数

	// 叠猫猫失败
	stackErr struct {
		ctx             *zero.Ctx // 上下文
		k               *meow     // 失败的猫猫
		l               int       // 叠猫猫队列高度
		n               int       // 造成别的猫猫退出的数量
		r               result    // 退出原因
		strings.Builder           // 错误内容
	}
*/
func stack(ctx *zero.Ctx, k *meow, l, n int, r result) *stackErr {
	go setCard(ctx, l-n)
	return &stackErr{
		ctx: ctx,
		k:   k,
		l:   l,
		n:   n,
		r:   r,
	}
}
