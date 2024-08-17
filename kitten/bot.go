package kitten

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"reflect"
	"slices"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type Item byte // 对上下文的检查类型

const (
	noEvent      = `非消息的上下文中获取的 bot 实例无 *Event，不可使用`
	Private      = `private`
	Group        = `group`
	Guild        = `guild`
	Caller  Item = iota // APICaller
	Event               // *Event
)

// Restart 重启 systemd 服务
func Restart(s string) {
	output, err := exec.Command(`sudo`, `systemctl`, `restart`, s).Output()
	if nil != err {
		Error(`重启 systemd 服务发生错误喵！`, string(output))
	}
}

// SendMessage 向 QQ（负数代表群聊）发送消息
func (u *QQ) SendMessage(ctx *zero.Ctx, message any) int64 {
	if u.IsQQ() {
		return ctx.SendPrivateMessage(u.Int(), message)
	}
	return ctx.SendGroupMessage(-u.Int(), message)
}

/*
SendText 发送文本

lf 控制群聊的 @ 后是否换行

非消息的上下文中获取的 bot 实例无事件，不可使用
*/
func SendText(ctx *zero.Ctx, lf bool, text ...any) message.MessageID {
	if !CheckCtx(ctx, Event) || !CheckCtx(ctx, Caller) {
		// 没有事件或 APICaller ，无法发送
		Warn(text...)
		return message.NewMessageIDFromInteger(0)
	}
	switch atUser := message.At(ctx.Event.UserID); ctx.Event.DetailType {
	case Private:
		return ctx.Send(Text(text...))
	case Group, Guild:
		if lf {
			return ctx.SendChain(atUser, Text("\n"), Text(text...))
		}
		fallthrough
	default:
		return ctx.SendChain(atUser, Text(text...))
	}
}

/*
SendTextOf 发送格式化文本

lf 控制群聊的 @ 后是否换行

非消息的上下文中获取的 bot 实例无事件，不可使用
*/
func SendTextOf(ctx *zero.Ctx, lf bool, format string, a ...any) message.MessageID {
	if !CheckCtx(ctx, Event) || !CheckCtx(ctx, Caller) {
		// 没有事件或 APICaller ，无法发送
		Warnf(format, a...)
		return message.NewMessageIDFromInteger(0)
	}
	switch atUser := message.At(ctx.Event.UserID); ctx.Event.DetailType {
	case Private:
		return ctx.Send(TextOf(format, a...))
	case Group, Guild:
		if lf {
			return ctx.SendChain(atUser, Text("\n"), TextOf(format, a...))
		}
		fallthrough
	default:
		return ctx.SendChain(atUser, TextOf(format, a...))
	}
}

/*
SendMessage 发送消息

lf 控制群聊的 @ 后是否换行

非消息的上下文中获取的 bot 实例无事件，不可使用
*/
func SendMessage(ctx *zero.Ctx, lf bool, m ...message.MessageSegment) message.MessageID {
	if !CheckCtx(ctx, Event) || !CheckCtx(ctx, Caller) {
		// 没有事件或 APICaller ，无法发送
		Warn(m)
		return message.NewMessageIDFromInteger(0)
	}
	switch messageChain := []message.MessageSegment{message.At(ctx.Event.UserID)}; ctx.Event.DetailType {
	case Private:
		return ctx.Send(m)
	case Group, Guild:
		if lf {
			return ctx.SendChain(append(append(messageChain, Text("\n")), m...)...)
		}
		fallthrough
	default:
		return ctx.SendChain(append(messageChain, m...)...)
	}
}

/*
SendWithImage 发送带有自定义图片的文字消息

非消息的上下文中获取的 bot 实例无事件，不可使用
*/
func SendWithImage(ctx *zero.Ctx, name core.Path, text ...any) message.MessageID {
	img, err := imagePath.Image(name)
	if nil != err {
		return SendText(ctx, true, err)
	}
	return SendMessage(ctx, true, img, Text(text...))
}

/*
SendWithImageOf 发送带有自定义图片的格式化文字消息

非消息的事件中获取的 bot 实例可能无效
*/
func SendWithImageOf(ctx *zero.Ctx, name core.Path, format string, a ...any) message.MessageID {
	img, err := imagePath.Image(name)
	if nil != err {
		return SendText(ctx, true, err)
	}
	return SendMessage(ctx, true, img, TextOf(format, a...))
}

// SendWithImageFail 发送带有失败图片的文字消息，非消息的事件中获取的 bot 实例可能无效
func SendWithImageFail(ctx *zero.Ctx, text ...any) message.MessageID {
	return SendWithImage(ctx, `no.png`, text...)
}

/*
SendWithImageFailOf 发送带有失败图片的格式化文字消息

非消息的事件中获取的 bot 实例可能无效
*/
func SendWithImageFailOf(ctx *zero.Ctx, format string, a ...any) message.MessageID {
	return SendWithImageOf(ctx, `no.png`, format, a...)
}

/*
DoNotKnow 喵喵不知道哦

非消息的事件中获取的 bot 实例可能无效
*/
func DoNotKnow(ctx *zero.Ctx) message.MessageID {
	return SendWithImageOf(ctx, `哈——？.png`, `%s不知道哦`, zero.BotConfig.NickName[0])
}

/*
GetObject 获取发送对象

返回正整数代表私聊，返回负整数代表群聊

返回默认值 0 代表不支持的对象
（非消息的事件中获取的 bot 实例 | 频道）
*/
func GetObject(ctx *zero.Ctx) QQ {
	if !CheckCtx(ctx, Event) {
		// 没有事件，无法获取
		return 0
	}
	switch ctx.Event.DetailType {
	case Private:
		return QQ(ctx.Event.UserID)
	case Group:
		return QQ(-ctx.Event.GroupID)
	default:
		return 0
	}
}

// CheckCtx 检查事件上下文的某个项目是否有效且不为空
func CheckCtx(ctx *zero.Ctx, i Item) bool {
	switch i {
	case Caller:
		c := reflect.ValueOf(ctx).Elem().FieldByName(`caller`)
		return c.IsValid() && !c.IsNil()
	case Event:
		if nil == ctx.Event {
			// 非消息的事件，直接返回
			Info(noEvent)
			return false
		}
	default:
		// 检查了错误的项目
		return false
	}
	return true
}

// GetSth 获取事件上下文中的字段
func GetSth[T any](ctx *zero.Ctx, name string) (t T) {
	f, ok := ctx.State[name]
	if !ok {
		return
	}
	t, _ = f.(T)
	return
}

// GetArgs 获取事件上下文中的参数
func GetArgs(ctx *zero.Ctx) string {
	return GetSth[string](ctx, `args`)
}

// GetArgsSlice 获取事件上下文中的参数切片
func GetArgsSlice(ctx *zero.Ctx) []string {
	return slices.DeleteFunc(strings.Split(GetArgs(ctx), ` `),
		func(s string) bool {
			return `` == s
		})
}

// GetImageURL 获取事件上下文中的图片链接
func GetImageURL(ctx *zero.Ctx) []string {
	return GetSth[[]string](ctx, `image_url`)
}

// ScanQRCode 扫描二维码
func ScanQRCode(name string) (fmt.Stringer, error) {
	var (
		msg = Image(name)
		n   = core.FilePath(`data`, `zbp`, `code.png`)
	)
	if _, err := core.GetImage(msg.Data[`file`], n); nil != err {
		return core.Path(msg.Data[`file`]), err
	}
	imgfile, err := os.Open(n.String())
	if nil != err {
		return core.Path(msg.Data[`file`]), err
	}
	defer imgfile.Close()
	img, _, err := image.Decode(imgfile)
	if nil != err {
		return nil, err
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if nil != err {
		return bmp, err
	}
	return qrcode.NewQRCodeReader().DecodeWithoutHints(bmp)
}
