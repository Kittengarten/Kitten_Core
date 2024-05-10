package view

import (
	"fmt"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const alipayvoiceURL = "https://mm.cqu.cc/share/zhifubaodaozhang/mp3/%v.mp3" // 支付宝到账语音

// 发送支付宝到账语音
func sendAlipayVoice(ctx *zero.Ctx) {
	ctx.Send(message.Record(fmt.Sprintf(alipayvoiceURL, strings.TrimSpace(kitten.GetArgs(ctx)))))
}
