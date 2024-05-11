package stack2

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/vicanso/go-charts/v2"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 概率组
type chance struct {
	p float64 // 压坏概率
	f float64 // 摔下概率
	s float64 // 成功概率
}

// 分析叠猫猫
func (d *data) analysis(ctx *zero.Ctx) message.MessageID {
	if cockroach == globalLocation {
		return kitten.SendWithImageFail(ctx, cockroachDoNotAnalysis)
	}
	c, f, img, err := d.generateAnalysis(ctx)
	if nil != err || !img {
		// 如果生成失败，或者不需要图片，什么也不做
		// 初始化时已经发送了相关信息
		return message.MessageID{}
	}
	core.RandomDelay(time.Second)
	if f {
		return d.analysisImage(ctx, c, true)
	}
	return d.analysisImage(ctx, c, false)
}

// 生成分析并发送文本
func (d *data) generateAnalysis(ctx *zero.Ctx) (c chance, flat, img bool, err error) {
	var (
		dn = slices.Clone(*d) // 克隆切片，防止对后续调用造成影响
		s  = dn.getStack()    // 获取叠猫猫队列
	)
	k, err := dn.pre(ctx) // 初始化自身
	if nil != err {
		// 如果不能加入，什么也不做
		// 初始化时已经发送了相关信息
		return
	}
	// 如果能加入
	// 猫娘少女和成年猫娘以上享有分析图片特权
	img = 猫娘少女 <= k.getTypeID(ctx)
	l := len(s) // 叠猫猫队列长度
	go setCard(ctx, l)
	if 0 == l {
		// 如果是空队列
		flat = true
		c.f = chanceFlat(k) // 平地摔概率
		c.s = 1 - c.f       // 成功概率
		sendTextOf(ctx, true, `【叠猫猫分析】
当前体重：　	%.1f kg
平地摔概率：	%.2f%%
成功概率：　	%.2f%%
%s`,
			itof(k.Weight),
			100*c.f,
			100*c.s,
			tipFlat(),
		)
		return
	}
	// 如果是非空队列
	sn := append(s, k)       // 用于压坏判定的队列
	c.p = sn.chancePressed() // 压坏概率
	gp := func(m meow) float64 {
		if 抱枕 >= k.getTypeID(ctx) || 幼年猫娘 <= m.getTypeID(ctx) {
			// 抱枕及以下的猫猫不会导致猫猫摔下去，直接在猫娘身上叠猫猫不会摔下去
			return 0
		}
		return k.chanceFall(m)
	}(s[l-1]) // 不压坏的情况下，摔下去的概率
	c.f = (1 - c.p) * gp // 摔下概率
	c.s = 1 - c.p - c.f  // 成功概率
	sendTextOf(ctx, true, `【叠猫猫分析】
猫堆高度：	%d
当前体重：	%.1f kg
压坏概率：	%.2f%%
摔下概率：	%.2f%%
成功概率：	%.2f%%
%s%s`,
		l,
		itof(k.Weight),
		100*c.p,
		100*c.f,
		100*c.s,
		func() string {
			cc := chanceClear(s, k, ctx)
			switch {
			case 0 == cc:
				return ``
			case 1e-4 <= cc:
				return fmt.Sprintf(`清空概率：	%.2f%%
`, 100*cc)
			default:
				return fmt.Sprintf(`清空概率：	%.2E
`, cc)
			}
		}(), tip(k.Weight, c),
	)
	return
}

// 计算清空猫堆的概率
func chanceClear(s data, k meow, ctx *zero.Ctx) float64 {
	var (
		sn   = append(s, k)       // 用于压坏判定的队列
		p, f = 1.0, 1.0           // 每次的压坏、摔下概率
		p1   = sn.chancePressed() // 压坏概率
		l    = len(s)             // 猫堆高度
	)
	if 0 == l {
		// 如果猫堆本来就是空的，清空概率等于平地摔概率
		return chanceFlat(k)
	}
	var (
		cf = func(m meow) float64 {
			if 抱枕 >= k.getTypeID(ctx) || 幼年猫娘 <= m.getTypeID(ctx) {
				// 抱枕及以下的猫猫不会导致猫猫摔下去，直接在猫娘以上级别的身上叠猫猫不会摔下去
				return 0
			}
			return k.chanceFall(m)
		}(s[l-1]) // 在不压坏的情况下，摔下去的概率
		ff = true // 计算摔下概率是否判定叠入的猫猫
	)
	for range sn {
		if 0 == p {
			break
		}
		if 1 == len(sn) {
			// 最后一只猫猫（叠入的猫猫本身），则已经清空
			sn = nil
			break
		}
		p *= sn.chancePressed() // 每次的压坏概率
		if 猫娘萝莉 <= sn[0].getTypeID(ctx) && 2 < len(sn) {
			// 如果底座是猫娘萝莉以上，则不会继续压坏
			// 如果此时剩余的猫堆高度大于 1，则无法清空
			p = 0
			break
			// 如果此时剩余的猫堆高度等于 1，则刚好清空
		}
		sn = sn[1:] // 去除压坏的猫猫
	}
	for range s {
		if 0 == f {
			break
		}
		if ff {
			// 如果是判定叠入的猫猫
			f *= cf
		}
		l = len(s)
		if 0 == l {
			s = nil
			break
		}
		if !ff {
			// 如果不是判定叠入的猫猫
			f *= k.chanceFall(s[l-1]) // 在不压坏的情况下，每次摔下去的概率
		}
		ff = false // 从此必然不是判定叠入的猫猫
		k = s[l-1]
		if 猫娘少女 <= k.getTypeID(ctx) && 2 <= l {
			// 如果摔下去的是猫娘少女以上级别，则下方的猫猫不会继续摔下去
			f = 0
		}
		s = s[:l-1] // 去除摔下去的猫猫
	}
	return p + (1-p1)*f
}

// 平地摔小贴士
func tipFlat() string {
	t := []string{
		`不来一发喵？`,
		`我的平地摔 我的平地摔`,
	}
	return t[rand.IntN(len(t))]
}

// 叠猫猫小贴士，除平地摔以外
func tip(w int, c chance) string {
	t := make([]string, 0, 64)
	if mapMeow[猫娘少女].weight <= w {
		t = append(t, []string{
			`哪有老虎一直饿！`,
			`嗷呜！`,
			`吃掉以后就能永远在一起了`,
			`桀桀桀，美味的小猫咪！
生来就是要被老虎吃掉的`,
			`大吉大利
今天吃猫`,
		}...,
		)
	}
	if 1 >= w {
		t = append(t, []string{
			`绒布球也是可以翻身的喵`,
			`惹不起，惹不起`,
			`今天对我爱理不理
明天让你高攀不起`,
		}...,
		)
	}
	if 10 > w {
		t = append(t, []string{
			`会长大的`,
			`小猫咪无所畏惧！`,
			`小猫咪勇往直前！`,
			`有我有你，一鼓作气！`,
			`小小的也很可爱呢`,
			`像你这样的小猫
生来就是要摔成绒布球的`,
		}...,
		)
	}
	if 0.5 <= c.p {
		t = append(t, []string{
			`是猪咪`,
			`加冕你为猪咪王`,
			`该减肥了`,
			`压猫猫！`,
			`请勿给猪染色`,
			`生存还是毁灭，你别无选择。`,
			`该毁灭了，猫堆。`,
			`让猫堆感受痛楚！`,
			`崩塌吧，旧猫堆！`,
			`所有，或者一无所有。`,
			`叠猫日久，当建奇功！
偷渡底座，直取猫堆！`,
		}...,
		)
	}
	if 0.5 <= c.f {
		t = append(t, []string{
			`哪有猫猫天天摔！`,
			`叠猫猫要笑着叠`,
			`我这么可爱，还要受这么大委屈`,
			`底座啊，我必偿还`,
			`猫堆啊，我已归来`,
			`床头叠上床尾摔
猫堆没有隔夜猫`,
			`喜欢是讨厌，讨厌就是喜欢
纯爱真的是很麻烦的东西`,
			`因为我已触碰过天空`,
			`英雄……可不应该……露出悲伤的……表情……
你还是……笑起来……最棒了……`,
			`人们为何选择沉睡？我想……
是因为害怕从「梦」中醒来。`,
			`你不要过来呀！`,
			`别让我掉下去，别让我掉下去～`,
			`菜，就多叠
摔不起，就别叠
以前是以前
现在是现在
你要是一直拿以前当现在
你怎么不和小奶猫比啊`,
			`爬得越高
摔得越疼`,
		}...,
		)
	}
	if 0.2 > c.s {
		t = append(t, []string{
			`进不去，怎么想都进不去吧？`,
			`这是一场豪赌，朋友`,
			`你能面对真正的内心吗？`,
			`已经没什么好害怕的了`,
			`那样的事，我决不允许`,
			`成为人类，就意味着隐藏秘密，经历痛苦，与孤独相伴，即便如此你也愿意吗？`,
			`你被强化了，快上！`,
			`不成功，便成仁！`,
			`为什么……为什么要演奏《春日影》！`,
			`你这个人，满脑子都只想着自己呢`,
			`什么时候都叠只会害了你`,
			`输了可不好玩
所以我从来都不会输`,
		}...,
		)
	}
	if 0.5 <= c.s && 0.8 > c.s {
		t = append(t, []string{
			`勇敢猫猫
不怕困难`,
			`就是这么自信`,
			`叠得要快，姿势要帅`,
			`做我的猫`,
			`你不叠，有的是猫猫叠`,
			`不要以为这样就赢了`,
			`在等什么呢`,
			`《冠军晚餐•猫的摇篮》`,
			`狂风呼啸着，这使你充满了决心`,
			`非常好猫堆，使我猫咪胖胖`,
			`随蝴蝶一起消散吧，旧日的幻影！`,
			`「来呀，来呀，花园的梦，森林的记忆…」
「来呀，来呀，不返的风，不逆流的水…」
「来呀，来呀，甜美的梦与苦涩的回味。」
「送别吧，让我们：」
「老去的落叶，胀满的果实…」
「淡去的好梦，谢落的花朵。」
「等待呀，让我们：」
「雨季归来，草木欢畅…」
「石榴歌唱，苹果鼓掌。」`,
			`「…黄金与白银、日轮与月镜相映的颜色，就是她们的友谊。」`,
			`「但愿新的梦想永远不被无留陀侵蚀。但愿旧的故事与无留陀一同被忘却。」
「但愿绿色的原野、山丘永远不变得枯黄。但愿溪水永远清澈，但愿鲜花永远盛开。」
「挚友将再次同行于茂密的森林中。一切美好的事物终将归来，一切痛苦的记忆也会远去，就像溪水净化自己，枯树绽出新芽。」`,
			`「不要输给风，不要输给雨。不输给风雪，也不输给炎夏…」`,
			`「我做了一个很长很长的梦…
「人们手握着手转圈。贤者与愚者，舞女与勇士，人偶与神像…
「大家的欢舞里蕴藏着宇宙的一切。『生命』一直都是目的，『智慧』才是手段。」`,
			`给岁月以文明，而不是给文明以岁月。`,
			`不 没关系
我的愿望是消灭所有的魔女
如果那真的实现了的话
我也是——
再没有绝望的必要了！`,
		}...,
		)
		if 1 > math.Pow(math.E, math.E)*rand.Float64() {
			t = append(t, `我们都栖息在智慧之树下
尝试阅读世界
从雨中读
尔后化身白鸟
攀上枝头
终于衔住了至关重要的那一片树叶
曾经
我是世上唯一能做梦的个体
在我的梦里
所有人入夜后也都会进入梦乡
人们的脑海中飘出奇思妙想
有些滚落地面
有些浮到天上
连接成一片万分夺目的网
三千世界之中
又有小小世界
所有命运
皆在此间沸腾
我逐渐明白
这些不可被描述
而又恒久变化之物
才是世间最深奥的东西
才能彻底驱逐那些疯狂
唯有梦
才能将意识
从最深沉的黑暗中唤醒
亦是求解之人
以世人之梦挽救世界
曾是属于我的答案
你们也寻到了属于自己的答案
我会将所有的梦
归还世人
得享美梦`)
		}
	}
	if 0 == len(t) {
		t = append(t, []string{
			`想听个故事吗`,
			`好想一天都缩在被窝里`,
			`你想去哪里呀`,
			`可以要个抱抱吗`,
			`你怎么傻乎乎的`,
			`如果你愿意的话`,
			`接下来你是不是要凶我了`,
			`会永远在一起的`,
			`又不是小孩子了`,
			`那可是个好消息`,
			`所以不会受伤`,
			`需要特别服务吗`,
			`不会让她等太久的`,
			`我就在这里哦`,
			`请让我来帮助你`,
			`生活真不容易呢`,
			`不会丢下我的吧`,
			`都是女孩子，没关系的`,
			`是不是想想就觉得好浪漫呀`,
			`一定会非常可爱的`,
			`奇迹和魔法都是存在的`,
			`能摘到星星真是太好了`,
			`天上可不会掉馅饼`,
			`那里当然会更加温暖的`,
			`这种事绝对很奇怪啊`,
			`一定会让姐姐满意的哦`,
			`总之就是超级可爱`,
			`这种东西小孩子可不能乱碰`,
			`旅途真是太美妙了`,
			`就会觉得非常幸福咯`,
			`大概算是很好养活的那种`,
			`如果能多带去一片温柔的话`,
			`我才没有那么好骗`,
			`就连风似乎也变得软绵绵的`,
			`今天都在哪里玩呀`,
			`糖纸上闪亮的是什么呢`,
			`比想象中还要近哦`,
			`快夸我快夸我`,
			`悄悄秘秘、小心翼翼`,
			`聪明可以用金币购买的吧`,
			`要一起出去走走吗`,
			`这世界从来不曾完美无瑕`,
			`希望能梦到温柔的地方`,
			`祝你早日寻到自己的宝石`,
			`别以为我不在，我随时都在`,
			`让大家做一个好梦吧`,
			`请允许我讲一个故事吧`,
		}...,
		)
	}
	return t[rand.IntN(len(t))]
}

// 查看叠猫猫图片
func (d *data) analysisImage(ctx *zero.Ctx, c chance, flat bool) message.MessageID {
	if !kitten.CheckCtx(ctx, kitten.Event) || !kitten.CheckCtx(ctx, kitten.Caller) {
		// 没有事件或 APICaller ，无法发送
		return message.MessageID{}
	}
	values := func() []float64 {
		if flat {
			return []float64{c.f, c.s}
		}
		return []float64{c.p, c.f, c.s}
	}()
	charts.SetDefaultWidth(640)
	charts.SetDefaultHeight(360)
	p, err := charts.PieRender(
		values,
		charts.TitleOptionFunc(charts.TitleOption{
			Text:    `叠猫猫分析`,
			Subtext: `概率`,
			Left:    charts.PositionCenter,
		}),
		charts.PaddingOptionFunc(charts.Box{
			Top:    20,
			Right:  20,
			Bottom: 20,
			Left:   20,
		}),
		charts.LegendOptionFunc(charts.LegendOption{
			Orient: charts.OrientVertical,
			Data: func() []string {
				if flat {
					return []string{`平地摔`, `成功`}
				}
				return []string{`压坏`, `摔下`, `成功`}
			}(),
			Left: charts.PositionLeft,
		}),
		charts.PieSeriesShowLabel(),
	)
	if nil != err {
		return sendWithImageFail(ctx, err)
	}
	buf, err := p.Bytes()
	if nil != err {
		return sendWithImageFail(ctx, err)
	}
	path := core.FilePath(imagePath, `temp.png`)
	if err = path.Write(buf); nil != err {
		return sendWithImageFail(ctx, err)
	}
	img, err := imagePath.Image(`temp.png`)
	if nil != err {
		return sendWithImageFail(ctx, err)
	}
	return ctx.Send(img)
}
