package view

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const waifu = `https://www.thiswaifudoesnotexist.net/example-%d.jpg` // AI 随机老婆

// 发送 AI 随机老婆
func sendWaifu(ctx *zero.Ctx) message.MessageID {
	return kitten.SendMessage(ctx, true, message.Image(fmt.Sprintf(waifu, rand.IntN(100001))))
}

// 从 ctx 参数中的 URL 发送图片
func sendImage(ctx *zero.Ctx) {
	img := kitten.GetArgs(ctx)
	if !zero.SuperUserPermission(ctx) && strings.Contains(`file://`, img) {
		kitten.SendWithImageFail(ctx, `权限不足喵！`)
		return
	}
	kitten.SendMessage(ctx, true, message.Image(img))
}

// 扫码
func scan(ctx *zero.Ctx) {
	if zero.HasPicture(ctx) {
		// 如果提供了图片，直接使用
		scanQRCode(ctx)
		return
	}
	// 如果没有提供图片，从链接解析
	img := kitten.GetArgs(ctx)
	if `` == img || !zero.SuperUserPermission(ctx) && strings.Contains(`file://`, img) {
		// 需要一张图片
		if zero.MustProvidePicture(ctx) {
			scanQRCode(ctx)
			return
		}
		kitten.SendWithImageFail(ctx, `没有收到图片，命令已过期喵！`)
		return
	}
	s, err := kitten.ScanQRCode(img)
	if nil != err {
		kitten.SendWithImageFail(ctx, `扫描失败喵！`, err)
		return
	}
	kitten.SendText(ctx, true, s)
}

// 从事件上下文的消息所附带的图片中扫描二维码并发送结果（支持多张图片）
func scanQRCode(ctx *zero.Ctx) message.MessageID {
	imgs := kitten.GetImageURL(ctx)
	if 0 == len(imgs) {
		return kitten.SendWithImageFail(ctx, `未获取到图片的链接喵！`)
	}
	r := make([]string, len(imgs), len(imgs))
	for i, img := range imgs {
		s, err := kitten.ScanQRCode(img)
		if nil != err {
			r[i] = err.Error()
			continue
		}
		r[i] = s.String()
	}
	return kitten.SendText(ctx, true, strings.Join(r, "\n"))
}
