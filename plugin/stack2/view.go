package stack2

import (
	"fmt"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/vicanso/go-charts/v2"

	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 查看叠猫猫
func (d *data) view(ctx *zero.Ctx, all bool) {
	var (
		s = d.getStack() // 获取叠猫猫队列
		v any            // 内容
	)
	if all {
		// 查看全部
		v = &s
	} else {
		// 查看省略版
		v = s.Str()
	}
	sendTextOf(ctx, true, `【叠猫猫队列】
现在有 %d 只猫猫
总重量为 %.1f kg
————%s`,
		len(s),
		itof(s.totalWeight()),
		v)
	go setCard(ctx, len(s))
}

// 初始化字体
func initFont() (err error) {
	// 获取字体数据
	buf, err := core.FilePath(text.GlowSansFontFile).Read()
	if nil != err {
		return
	}
	// 安装字体
	if err = charts.InstallFont(`glow`, buf); nil != err {
		return
	}
	// 加载字体
	font, err := charts.GetFont(`glow`)
	if nil != err {
		return
	}
	// 设置默认字体
	charts.SetDefaultFont(font)
	return
}

// 查看叠猫猫图片
func (d *data) viewImage(ctx *zero.Ctx) message.MessageID {
	var (
		s = d.getStack() // 获取叠猫猫队列
		l = len(s)       // 叠猫猫队列长度
	)
	if 2 > l {
		return message.MessageID{}
	}
	var (
		values = make([][]float64, 1, 1) // 叠猫猫图示数据
		str    = make([]string, l, l)    // 叠猫猫图示文字
	)
	values[0] = make([]float64, l, l) // 初始化二维数组
	for h, k := range s {
		values[0][h] = itof(k.Weight)
		origin := func() string {
			if cockroach == globalLocation {
				return fmt.Sprintf(`【%s】翼展 %.1f cm`, mapMeow[k.getTypeID(globalCtx)].str, itof(k.Weight))
			}
			return fmt.Sprintf(`%s（%d）%.1f kg%s`,
				k.TitleCardOrNickName(globalCtx), k.Int(), itof(k.Weight),
				l10nReplacer(globalLocation).Replace(mapMeow[k.getTypeID(ctx)].str))
		}
		str[h] = strings.ReplaceAll(origin(), `	`, ``)
	}
	p, err := setChart(values, str, l)
	if nil != err {
		return sendWithImageFail(ctx, err)
	}
	return sendImage(ctx, p)
}

// 设置图表
func setChart(v [][]float64, s []string, l int) (p *charts.Painter, err error) {
	charts.SetDefaultWidth(max(min(3840, 160+320*l), 960))
	charts.SetDefaultHeight(max(min(2160, 90*l), 270))
	return charts.HorizontalBarRender(
		v,
		charts.TitleTextOptionFunc(l10nReplacer(globalLocation).Replace(`叠猫猫队列`)),
		charts.PaddingOptionFunc(charts.Box{
			Top:    20,
			Right:  40,
			Bottom: 20,
			Left:   40,
		}),
		charts.LegendLabelsOptionFunc([]string{l10nReplacer(globalLocation).Replace(`体重（kg）`)}),
		charts.YAxisDataOptionFunc(s),
	)
}

// 生成并发送图片
func sendImage(ctx *zero.Ctx, p *charts.Painter) message.MessageID {
	buf, err := p.Bytes()
	if nil != err {
		return message.MessageID{}
	}
	path := core.FilePath(imagePath, `temp.png`)
	if err = path.Write(buf); nil != err {
		sendWithImageFail(ctx, err)
	}
	img, err := imagePath.Image(`temp.png`)
	if nil != err {
		return sendWithImageFail(ctx, err)
	}
	return ctx.Send(img)
}
