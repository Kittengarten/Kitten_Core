// Package stack2 叠猫猫 v2
package stack2

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName           = `stack2` // 插件名
	brief                      = `一起来玩叠猫猫 v2`
	dataFile                   = `data.yaml` // 叠猫猫数据文件
	cStack, cStackT0, cStackT1 = `叠`, `曡`, `疊`
	cMeow                      = `猫猫`
	cIn                        = `加入`
	cView                      = `查看`
	cAnalysis                  = `分析`
	cRank                      = `排行`
	cEat                       = `吃`
	cEatGPU                    = `抢`
	zako                       = `杂鱼.png`
)

var (
	// 全局上下文，仅用于猫猫的 String() 方法
	globalCtx *zero.Ctx
	// 可导出的读写锁，用于叠猫猫文件的并发安全
	Mu sync.RWMutex
)

func init() {
	if nil != err {
		kitten.Error(`叠猫猫配置文件错误喵！`, err)
		return
	}
	// 初始化字体
	if err := initFont(); nil != err {
		kitten.Error(`字体初始化错误喵！`, err)
	}

	// 叠猫猫
	engine.OnCommandGroup([]string{cStack, cStackT0, cStackT1}).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.User)).
		Limit(kitten.GetLimiter(kitten.GroupFast)).
		Handle(stackExe)

	// 吃猫猫
	engine.OnCommandGroup([]string{cEat, cEatGPU}).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.User)).
		Limit(kitten.GetLimiter(kitten.GroupFast)).
		Handle(eatExe)
}

// 设置全局地区标记位
func setGlobalLocation(s string) bool {
	switch {
	case strings.ContainsAny(s, `狐狸`):
		globalLocation = fox // 狐狐
		return true
	case strings.Contains(s, `显卡`):
		globalLocation = gpu // 显卡
		return true
	case strings.ContainsAny(s, `蟑螂`),
		strings.Contains(s, `小强`),
		strings.Contains(s, `大蠊`):
		globalLocation = cockroach // 蟑螂
		return checkCockroachDate()
	case strings.ContainsAny(s, `猫虎`):
		fallthrough // 猫猫
	default:
		globalLocation = cat // 默认叠猫猫
		return true
	}
}

// 叠猫猫执行逻辑
func stackExe(ctx *zero.Ctx) {
	args := slices.DeleteFunc(strings.Split(kitten.GetArgs(ctx), ` `),
		func(s string) bool {
			return `` == s
		})
	if 2 != len(args) {
		kitten.SendWithImageFailOf(ctx, `本命令参数数量：2
%s%s%s %s|%s|%s|%s
传入的参数数量：%d
参数数量错误，请用半角空格隔开各参数喵！`,
			p, cStack, cMeow, cIn, cView, cAnalysis, cRank,
			len(args))
		return
	}
	if !setGlobalLocation(args[0]) {
		// 设置全局地区标记位，如当前活动未开放则返回
		kitten.SendWithImageFail(ctx, `当前活动未开放喵！`)
		return
	}
	globalCtx = ctx
	switch args[1] {
	case cIn:
		// 如果是加入，需要写锁
		Mu.Lock()
		defer Mu.Unlock()
	case cView, cAnalysis, cRank:
		// 如果是查看、分析、排行，需要读锁
		fallthrough
	default:
		// 默认显示帮助文本，需要读锁
		Mu.RLock()
		defer Mu.RUnlock()
	}
	d, err := core.Load[data](dataPath, core.Empty)
	if nil != err {
		sendWithImageFail(ctx, `加载叠猫猫数据文件时发生错误喵！`, err)
		return
	}
	switch args[1] {
	case cIn:
		d.in(ctx)
		selfEat(ctx, d, p)
		core.RandomDelay(time.Second)
		selfIn(ctx, d, p)
	case cView:
		d.view(ctx, zero.UserOrGrpAdmin(ctx))
		core.RandomDelay(time.Second)
		d.viewImage(ctx)
	case cAnalysis:
		d.analysis(ctx)
		selfAnalysis(ctx, d, p)
		selfEat(ctx, d, p)
		core.RandomDelay(time.Second)
		selfIn(ctx, d, p)
	case cRank:
		d.rank(ctx)
		selfRank(ctx, d, p)
	default:
		var (
			u    = ctx.Event.UserID
			k, i = d.getMeow(u)
			w    int
		)
		if -1 != i {
			w = k.Weight
		} else {
			qq := kitten.QQ(u)
			w = len(qq.TitleCardOrNickName(ctx))
		}
		sendText(ctx, true, strings.NewReplacer(
			`(抱枕突破所需体重/当前体重)`,
			fmt.Sprintf(` %.2f%% `, 100*chanceFlat(k)),
			`N(0, 体重²)`,
			fmt.Sprintf(`N(0, (%s)²)`, core.ConvertTimeDuration(
				time.Hour*time.Duration(stackConfig.GapTime*w)/10)),
		).Replace(help))
		if 小老虎 <= k.getTypeID(ctx) {
			// 如果是老虎，发送吃猫猫帮助文本
			core.RandomDelay(time.Second)
			sendText(ctx, true, strings.NewReplacer(
				`N(0, (e*体重)²)`,
				fmt.Sprintf(`N(0, (%s)²)`, core.ConvertTimeDuration(
					time.Duration(float64(stackConfig.GapTime)*float64(time.Hour)*math.E*itof(w)))),
			).Replace(helpEat))
		}
	}
}

/*
叠猫猫尝试加入前的初始化，返回叠入的猫猫

如果不用于叠入，则需要克隆切片

错误已经打印，无需重复打印
*/
func (d *data) pre(ctx *zero.Ctx) (meow, error) {
	var (
		u = ctx.Event.UserID // 叠入猫猫的 QQ
		w int                // 叠入猫猫的体重
		b time.Duration      // 剩余的休息时间
	)
	if slices.ContainsFunc(*d, func(k meow) bool {
		b = k.Time.Sub(time.Unix(ctx.Event.Time, 0))
		w = k.Weight
		return u == k.Int() && !k.Status && 0 < b
	}) {
		err := needRest(b, w)
		if sid.Int() == u {
			kitten.Weight = w
			return meow{}, err
		}
		sendWithImageFail(ctx, err)
		return meow{}, err
	}
	if slices.ContainsFunc(*d, func(k meow) bool { return u == k.Int() && k.Status }) {
		err := alreadyJoined()
		if sid.Int() == u {
			kitten.Weight = w
			return meow{}, err
		}
		sendWithImageFail(ctx, err)
		return meow{}, err
	}
	var (
		qq   = kitten.QQ(u)                // 叠入的猫猫 QQ
		name = qq.TitleCardOrNickName(ctx) // 叠入的猫猫名称
		i    int                           // 叠入的猫猫下标
	)
	k, i := d.getMeow(u) // 获取叠入的猫猫及其下标，如果不用于叠入，则需要克隆切片
	if -1 == i {
		// 如果是首次叠猫猫
		k = meow{
			Name:   name,
			Weight: max(1, len(name)),
			Time:   time.Unix(ctx.Event.Time, 0),
		}
		k.Set(u)
		return k, nil
	}
	// 如果是已经存在的猫猫，更新其名称
	k.Name = name
	return k, nil
}

/*
清空猫堆特效

根据是否清空猫堆，添加提示语

l 为队列高度，n 为结果，w 为叠猫猫前的体重
*/
func doClear(l, n int, w int, k *meow, r *strings.Builder) {
	r.WriteByte('\n')
	if l == n {
		// 如果清空了猫堆
		if clear(k) {
			r.WriteString(`你触发了清空猫堆的特效！`)
		} else {
			r.WriteString(`你清空了猫堆，但没有发生特别的事情。`)
		}
		r.WriteByte('\n')
	}
	// 如果没有清空猫堆
	if k.Weight == w {
		r.WriteString(fmt.Sprintf(`你的体重为 %.1f kg 不变。`, itof(w)))
		r.WriteByte('\n')
		return
	}
	r.WriteString(fmt.Sprintf(`你的体重由 %.1f kg 变为 %.1f kg。`, itof(w), itof(k.Weight)))
	r.WriteByte('\n')
}

/*
执行叠猫猫，k 为叠入的猫猫

错误已经打印，无需重复打印
*/
func (d *data) doStack(ctx *zero.Ctx, k *meow) error {
	*d = d.getStack() // 正在叠猫猫的队列
	var (
		dr = slices.Clone(*d) // 叠猫猫队列的克隆
		l  = len(dr)          // 叠猫猫队列高度
	)
	if d.checkFlat(*k) {
		// 如果平地摔
		err := stack(ctx, k, l, 0, flat)
		sendWithImage(ctx, core.Path(zako), err)
		return err
	}
	if p := d.pressResult(ctx, *k); 0 != p {
		// 压坏了别的猫猫
		var (
			err = stack(ctx, k, l, p, press)
			e   = dr[:p]
		)
		sendWithImage(ctx, core.Path(zako), err, &e)
		return err
	}
	// 如果没有猫猫被压坏，叠猫猫初步成功
	if f := d.fallResult(ctx, *k); 0 != f {
		// 摔坏了别的猫猫
		var (
			err = stack(ctx, k, l, f, fall)
			e   = dr[l-f:]
		)
		sendWithImage(ctx, core.Path(zako), err, &e)
		return err
	}
	// 如果没有摔坏猫猫，叠猫猫成功
	k.Status = true
	sendTextOf(ctx, true, `叠猫猫成功，目前处于队列中第 %d 位喵～
你的当前体重为 %.1f kg。`,
		1+l,
		itof(k.Weight))
	go setCard(ctx, 1+l)
	return nil
}

/*
加入叠猫猫，当且仅当叠猫猫失败时返回的是 *stackErr

错误已经打印，无需重复打印
*/
func (d *data) in(ctx *zero.Ctx) error {
	// 初始化自身
	k, err := d.pre(ctx)
	if nil != err {
		return err
	}
	// 未在叠猫猫的队列
	dn := d.getNoStack()
	// 执行叠猫猫
	e := d.doStack(ctx, &k)
	// 合并当前未叠猫猫与叠猫猫的队列，将叠入的猫猫追加入切片中
	*d = slices.Concat(dn, *d, data{k})
	// 存储叠猫猫数据
	if err := core.Save(dataPath, d); nil != err {
		sendWithImageFail(ctx, `存储叠猫猫数据时发生错误喵！`, err)
		return err
	}
	return e
}

// 获取并返回叠猫猫队列
func (d *data) getStack() data {
	// 删除不在叠猫猫中的猫猫，得到叠猫猫队列
	return slices.DeleteFunc(slices.Clone(*d), func(k meow) bool { return !k.Status })
}

// 获取并返回不在叠猫猫的队列
func (d *data) getNoStack() data {
	// 删除叠猫猫中的猫猫，得到不在叠猫猫的队列
	return slices.DeleteFunc(slices.Clone(*d), func(k meow) bool { return k.Status })
}

/*
提取猫猫及其下标，会从切片中删除提取的猫猫

无此猫猫则返回空结构体及 -1
*/
func (d *data) getMeow(u int64) (meow, int) {
	i := slices.IndexFunc(*d, func(k meow) bool { return u == k.Int() })
	if -1 == i {
		return meow{}, i
	}
	m := (*d)[i]
	*d = slices.Delete(*d, i, 1+i)
	return m, i
}

/*
String 实现 fmt.Stringer

从叠猫猫队列生成完整字符串（开头有一次换行）
*/
func (d *data) String() string {
	// 克隆一份防止修改源数据
	dn := slices.Clone(*d)
	// 按“后来居上”排列叠猫猫队列
	slices.Reverse(dn)
	var s strings.Builder
	s.Grow(32 * len(dn))
	for _, k := range dn {
		s.WriteByte('\n')
		s.WriteString(k.String())
	}
	return s.String()
}

/*
从叠猫猫队列生成省略过的字符串

队列高度不超过 20 时，无需省略
*/
func (d *data) Str() string {
	var (
		dn = slices.Clone(*d) // 克隆一份防止修改源数据
		l  = len(dn)          // 叠猫猫队列高度
		s  strings.Builder
		ok bool
	)
	s.Grow(32 * min(l, 20))
	// 按“后来居上”排列叠猫猫队列
	slices.Reverse(dn)
	for i, k := range dn {
		if 20 < l && 5 <= i && i < l-5 {
			// 当高度 > 20 时，跳过中间的猫猫，只取上下 5 只
			if ok {
				continue
			}
			s.WriteByte('\n')
			s.WriteString(`…………`)
			s.WriteByte('\n')
			for range l - 10 {
				s.WriteRune('🐱')
			}
			s.WriteByte('\n')
			s.WriteString(`…………`)
			ok = true
			continue
		}
		s.WriteByte('\n')
		s.WriteString(k.String())
	}
	return s.String()
}

// 获取全队列的总重量（0.1kg 数）
func (d *data) totalWeight() (w int) {
	for _, k := range *d {
		if core.MaxInt-k.Weight < w {
			// 防止溢出
			return core.MaxInt
		}
		w += k.Weight
	}
	return
}

// 获取最下方的猫猫被压坏的概率
func (d *data) chancePressed() float64 {
	// 压坏的概率
	if 1 >= len(*d) {
		// 如果只有一只猫猫或者没有猫猫，直接返回，避免下标越界
		return 0
	}
	a := (*d)[1:]
	return max(0, (float64(a.totalWeight())-math.E*float64((*d)[0].Weight))/
		float64(d.totalWeight()))
}

/*
检查最下方的猫猫是否被压坏

如果没有被压坏则返回 true
*/
func (d *data) checkPress() bool {
	return d.chancePressed() <= rand.Float64()
}

/*
获取被压坏猫猫的数量，并将被压坏的猫猫标记为未在叠猫猫

不含叠入的猫猫

正在叠猫猫的队列才能调用
*/
func (d *data) pressResult(ctx *zero.Ctx, k meow) int {
	var (
		s = append(*d, k) // 将叠入的猫猫纳入队列重量计算
		l = len(*d)       // 原队列高度
	)
	for i := range *d {
		n := &(*d)[i]
		if a := s[i:]; a.checkPress() {
			// 如果没有被压坏，则直接返回
			return i
		}
		// 去除压坏的猫猫
		exit(ctx, n, pressed, l-i)
		// 如果压坏的是猫娘萝莉，则不会继续压坏上方的猫猫
		if 猫娘萝莉 <= n.getTypeID(ctx) {
			return 1 + i
		}
	}
	return l
}

// 检查是否平地摔，正在叠猫猫的队列才能调用
func (d *data) checkFlat(k meow) bool {
	// 当叠猫猫队列为空， 抱枕突破所需体重/当前体重的概率平地摔
	return 0 == len(*d) && chanceFlat(k) > rand.Float64()
}

/*
获取叠猫猫失败摔下去猫猫的数量，并将摔下去的猫猫标记为未在叠猫猫

不含叠入的猫猫

正在叠猫猫的队列才能调用
*/
func (d *data) fallResult(ctx *zero.Ctx, k meow) int {
	// 初始猫猫数量
	l := len(*d)
	if 0 == l || 抱枕 >= k.getTypeID(ctx) || 幼年猫娘 <= (*d)[l-1].getTypeID(ctx) {
		// 抱枕及以下的猫猫不会导致猫猫摔下去，直接在猫娘以上级别的身上叠猫猫不会摔下去
		return 0
	}
	// 从队列的最上部开始遍历（后来居上）
	for i := range *d {
		// 下方的猫猫
		n := &(*d)[l-i-1]
		if k.checkFall(*n) {
			// 这只猫猫没有摔下去，直接返回
			return i
		}
		k = *n
		// 去除摔下去的猫猫
		exit(ctx, n, fall, l-i)
		if 猫娘少女 <= n.getTypeID(ctx) {
			// 如果摔下去的是猫娘少女以上级别，则下方的猫猫不会继续摔下去
			return 1 + i
		}
	}
	return l
}

/*
去除退出的猫猫 k，并使其进入休息，然后调整体重

t 为退出原因，h 为 摔下去的高度 | 压坏的猫猫总数 | 上方的猫猫总数 | 吃掉的猫猫总重量（0.1kg 数）
*/
func exit(ctx *zero.Ctx, k *meow, t result, h int) {
	// 去除
	k.Status = false
	// 休息
	r := float64(time.Hour) * float64(stackConfig.GapTime) * normal(itof(k.Weight))
	if eat == t {
		// 吃猫猫的休息时间为 e 倍
		r *= math.E
	}
	k.Time = time.Unix(ctx.Event.Time, 0).Add(max(
		time.Hour*time.Duration(stackConfig.MinGapTime),
		time.Duration(r)),
	)
	// 体重变化
	switch t {
	case flat:
		// 平地摔，体重变为 e 倍
		w := int(math.RoundToEven(math.E * float64(k.Weight)))
		k.Weight = max(w, -(1 + w))
	case fall:
		// 摔下去，体重 - 100g × 当前高度
		k.Weight = max(1, k.Weight-h)
	case press, pressed, eat:
		// 压坏了猫猫，体重 + 100g × 压坏的猫猫总数
		// 被压坏，体重 + 100g × 上方的猫猫总数
		// 吃猫猫，体重 + 吃掉的猫猫总重量（0.1kg 数）
		k.Weight = min(k.Weight, core.MaxInt-h) + h
	}
	// 被老虎吃掉，体重不变
}

// 清空猫堆的体重调整
func clear(k *meow) bool {
	if float64(mapMeow[抱枕].weight)/float64(k.Weight) <= rand.Float64() {
		return false
	}
	// 以抱枕突破所需体重/当前体重的概率，体重变为 e 倍
	w := int(math.RoundToEven(math.E * float64(k.Weight)))
	k.Weight = max(w, -(1 + w))
	return true
}
