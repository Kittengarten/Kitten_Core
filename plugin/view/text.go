package view

import (
	"encoding/json"
	"fmt"

	"github.com/Kittengarten/KittenAnno/wta"
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	jiTang  = `https://api.btstu.cn/yan/api.php?charset=utf-8&encode=text` // 鸡汤
	qingHua = `https://xiaobai.klizi.cn/API/other/wtqh.php`                // 情话
	kfc     = `https://api.jixs.cc/api/wenan-fkxqs/index.php`              // 疯狂星期四
	yiYan   = `https://v1.hitokoto.cn/?c=a&c=b&c=c&c=d&c=h&c=i`            // 动漫 漫画 游戏 文学 影视 诗词（一言）
)

// 发送网页内容，lf 控制内容是否换行
func send(ctx *zero.Ctx, url string, lf bool) {
	// 获取 HTTP 响应体，失败则返回
	data, err := core.GETData(url)
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	kitten.SendText(ctx, true, string(core.CleanAll(data, lf)))
}

// 发送一言
func sendYiYan(ctx *zero.Ctx) {
	var (
		rsp struct {
			Hitokoto string `json:"hitokoto"`
			From     string `json:"from"`
			FromWho  string `json:"from_who"`
		}
		// 获取 HTTP 响应体，失败则返回
		data, err = core.GETData(yiYan)
	)
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	if err := json.Unmarshal(data, &rsp); nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	kitten.SendText(ctx, true, rsp.Hitokoto+`
出自：`+rsp.From+func() string {
		if 0 == len(rsp.FromWho) {
			return ``
		}
		return `
作者：` + rsp.FromWho
	}(),
	)
}

// 返回世界树纪元
func getWTA() string {
	a, err := wta.GetAnno()
	if nil != err {
		kitten.Error(`报时失败喵！`, err)
		return `喵？`
	}
	return nickname + `报时：
日期：	` + a.DateStr() + `
时间：	` + a.String() + `
琴弦：	` + a.Chord() + `
花卉：	` + a.Flower() + `
` + a.ElementalAndImageryStr()
}

// 返回叠猫猫体重字符串
func weight() string {
	if 0 == kitten.Weight {
		return ``
	}
	return fmt.Sprintf(`	♥	叠猫猫体重：	%.1f kg`, float64(kitten.Weight)/10)
}
