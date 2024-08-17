package stack2

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/vicanso/go-charts/v2"

	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 查看叠猫猫
func (d *data) view(ctx *zero.Ctx, all bool) {
	s := d.getStack() // 获取叠猫猫队列
	go setCard(ctx, len(s))
	sendTextOf(ctx, true, `【叠猫猫队列】
现在有 %d 只猫猫
总重量为 %.1f kg
————%s`,
		len(s),
		itof(s.totalWeight()),
		func() any {
			if !all {
				// 查看省略版
				return s.Str()
			}
			// 查看全部（先发送前 50 条）
			sr := s[max(0, len(s)-50):]
			// 剩余部分
			s = s[:len(s)-len(sr)]
			return &sr
		}())
	for all && 0 < len(s) {
		core.RandomDelay(time.Second)
		// 发送剩余部分的前 50 条
		sr := s[max(0, len(s)-50):]
		// 剩余部分的剩余部分
		s = s[:len(s)-len(sr)]
		sendText(ctx, true, &sr)
	}
}

// 初始化字体
func initFont() error {
	// 获取字体数据
	buf, err := core.FilePath(text.GlowSansFontFile).ReadBytes()
	if nil != err {
		return err
	}
	// 安装字体
	const fontName = `glow`
	if err := charts.InstallFont(fontName, buf); nil != err {
		return err
	}
	// 加载字体
	font, err := charts.GetFont(fontName)
	if nil != err {
		return err
	}
	// 设置默认字体
	charts.SetDefaultFont(font)
	return nil
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
			return fmt.Sprintf(`%s（%d）%.1f %s %s`,
				k.TitleCardOrNickName(globalCtx), k.Int(), itof(k.Weight),
				l10nReplacer(globalLocation).Replace(`kg`),
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
func setChart(v [][]float64, s []string, l int) (*charts.Painter, error) {
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
	defer p.Close()
	if nil != err {
		return message.MessageID{}
	}
	path := core.FilePath(imagePath, `temp.png`)
	if _, err = path.Write(buf); nil != err {
		sendWithImageFail(ctx, err)
	}
	img, err := imagePath.Image(`temp.png`)
	if nil != err {
		return sendWithImageFail(ctx, err)
	}
	return ctx.Send(img)
}
