package kitten

import (
	"fmt"

	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

/*
TextOf 格式化构建 message.MessageSegment 文本

格式同 fmt.Sprintf
*/
func TextOf(format string, a ...any) message.MessageSegment {
	return message.Text(fmt.Sprintf(format, a...))
}

/*
SendTextOf 发送格式化文本

lf 控制群聊的 @ 后是否换行（非消息的事件中获取的 bot 实例可能无效）
*/
func SendTextOf(ctx *zero.Ctx, lf bool, format string, a ...any) {
	switch atUser, formattedText := message.At(ctx.Event.UserID), TextOf(format, a...); ctx.Event.DetailType {
	case `private`:
		ctx.Send(formattedText)
	case `group`, `guild`:
		if lf {
			ctx.SendChain(atUser, message.Text("\n"), formattedText)
			return
		}
		fallthrough
	default:
		ctx.SendChain(atUser, formattedText)
	}
}

/*
SendText 发送文本

lf 控制群聊的 @ 后是否换行（非消息的事件中获取的 bot 实例可能无效）
*/
func SendText(ctx *zero.Ctx, lf bool, text string) {
	switch atUser := message.At(ctx.Event.UserID); ctx.Event.DetailType {
	case `private`:
		ctx.Send(text)
	case `group`, `guild`:
		if lf {
			ctx.SendChain(atUser, message.Text("\n", text))
			return
		}
		fallthrough
	default:
		ctx.SendChain(atUser, message.Text(text))
	}
}

/*
SendMessage 发送消息

lf 控制群聊的 @ 后是否换行（非消息的事件中获取的 bot 实例可能无效）
*/
func SendMessage(ctx *zero.Ctx, lf bool, m ...message.MessageSegment) {
	switch messageChain := []message.MessageSegment{message.At(ctx.Event.UserID)}; ctx.Event.DetailType {
	case `private`:
		ctx.Send(m)
	case `group`, `guild`:
		if lf {
			ctx.SendChain(append(append(messageChain, message.Text("\n")), m...)...)
			return
		}
		fallthrough
	default:
		ctx.SendChain(append(messageChain, m...)...)
	}
}

// SendWithImage 发送带有自定义图片的文字消息
func SendWithImage(ctx *zero.Ctx, name Path, format string, a ...any) {
	img, err := imagePath.GetImage(name)
	if nil != err {
		SendTextOf(ctx, true, `%v`, err)
		return
	}
	SendMessage(ctx, true, img, TextOf(format, a...))
}

// SendWithImageFail 发送带有失败图片的文字消息
func SendWithImageFail(ctx *zero.Ctx, format string, a ...any) {
	img, err := imagePath.GetImage(`no.png`)
	if nil != err {
		SendTextOf(ctx, true, `%v`, err)
		return
	}
	SendMessage(ctx, true, img, TextOf(format, a...))
}

// DoNotKnow 喵喵不知道哦
func DoNotKnow(ctx *zero.Ctx) {
	img, err := imagePath.GetImage(`哈——？.png`)
	if nil != err {
		SendTextOf(ctx, true, `%v`, err)
		return
	}
	SendMessage(ctx, true, img, TextOf(`%s不知道哦`, zero.BotConfig.NickName[0]))
}

// GetTitleCardOrNickName 从 QQ 获取【头衔】群昵称或昵称
func (u QQ) GetTitleCardOrNickName(ctx *zero.Ctx) string {
	// 修剪后的昵称
	name := CleanAll(ctx.GetStrangerInfo(int64(u), true).Get(`nickname`).Str, false)
	if 0 >= ctx.Event.GroupID {
		// 不是群聊，直接返回昵称
		return name
	}
	// 是群聊，获取该 QQ 在群内的资料
	var (
		gmi   = ctx.GetThisGroupMemberInfo(int64(u), true) // 本群成员信息
		title = gmi.Get(`title`).Str                       // 头衔
	)
	if `` != title {
		// 如果头衔存在，则添加实心方头括号
		title = fmt.Sprintf(`【%s】`, title)
	}
	// 获取修剪后的群昵称
	if card := CleanAll(gmi.Get(`card`).Str, false); `` != card {
		// 如果不为空，返回【头衔】群昵称
		return fmt.Sprint(title, card)
	}
	// 返回【头衔】昵称
	return fmt.Sprint(title, name)
}

// （私有）获取信息
func (u QQ) getInfo(ctx *zero.Ctx) gjson.Result {
	return ctx.GetStrangerInfo(int64(u), true)
}

// IsAdult 是成年人
func (u QQ) IsAdult(ctx *zero.Ctx) bool {
	return 18 <= u.getInfo(ctx).Get(`age`).Int()
}

// IsFemale 是女性
func (u QQ) IsFemale(ctx *zero.Ctx) bool {
	return `female` == u.getInfo(ctx).Get(`sex`).String()
}
