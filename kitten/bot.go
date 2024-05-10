package kitten

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"reflect"

	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type Item byte // 对上下文的检查类型

const (
	noEvent                = `非消息的上下文中获取的 bot 实例无 *Event，不可使用`
	DetailTypePrivate      = `private`
	DetailTypeGroup        = `group`
	DetailTypeGuild        = `guild`
	Caller            Item = iota // APICaller
	Event                         // *Event
)

// Restart 重启 systemd 服务
func Restart(s string) {
	output, err := exec.Command(`sudo`, `systemctl`, `restart`, s).Output()
	if nil != err {
		Error(`重启 systemd 服务发生错误喵！`, string(output))
	}
}

/*
Text 构建 message.MessageSegment 文本

格式同 fmt.Sprint
*/
func Text(text ...any) message.MessageSegment {
	return message.Text(text...)
}

/*
TextOf 格式化构建 message.MessageSegment 文本

格式同 fmt.Sprintf
*/
func TextOf(format string, a ...any) message.MessageSegment {
	return message.Text(fmt.Sprintf(format, a...))
}

/*
SendText 发送文本

lf 控制群聊的 @ 后是否换行

非消息的上下文中获取的 bot 实例无事件，不可使用
*/
func SendText(ctx *zero.Ctx, lf bool, text ...any) message.MessageID {
	checkErr(text)
	if !CheckCtx(ctx, Event) || !CheckCtx(ctx, Caller) {
		// 没有事件或 APICaller ，无法发送
		Info(text...)
		return message.NewMessageIDFromInteger(0)
	}
	switch atUser := message.At(ctx.Event.UserID); ctx.Event.DetailType {
	case DetailTypePrivate:
		return ctx.Send(Text(text...))
	case DetailTypeGroup, DetailTypeGuild:
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
	checkErr(a)
	if !CheckCtx(ctx, Event) || !CheckCtx(ctx, Caller) {
		// 没有事件或 APICaller ，无法发送
		Infof(format, a...)
		return message.NewMessageIDFromInteger(0)
	}
	switch atUser := message.At(ctx.Event.UserID); ctx.Event.DetailType {
	case DetailTypePrivate:
		return ctx.Send(TextOf(format, a...))
	case DetailTypeGroup, DetailTypeGuild:
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
		Info(m)
		return message.NewMessageIDFromInteger(0)
	}
	switch messageChain := []message.MessageSegment{message.At(ctx.Event.UserID)}; ctx.Event.DetailType {
	case DetailTypePrivate:
		return ctx.Send(m)
	case DetailTypeGroup, DetailTypeGuild:
		if lf {
			return ctx.SendChain(append(append(messageChain, message.Text("\n")), m...)...)
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
	checkErr(text)
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
	checkErr(a)
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
GetObject 获取发送对象，admin 代表是否需要管理员权限

返回正整数代表群，返回负整数代表私聊

返回默认值 0 代表不支持的对象
（非消息的事件中获取的 bot 实例 | 频道）

返回 1 代表在群中无权限
*/
func GetObject(ctx *zero.Ctx, admin bool) int64 {
	if !CheckCtx(ctx, Event) {
		// 没有事件，无法获取
		return 0
	}
	switch ctx.Event.DetailType {
	case DetailTypePrivate:
		return -ctx.Event.UserID
	case DetailTypeGroup:
		if !admin || zero.AdminPermission(ctx) {
			return ctx.Event.GroupID
		}
		SendWithImageFail(ctx, `管理员才能使用喵！`)
		return 1
	case DetailTypeGuild:
		SendWithImageFail(ctx, `暂不支持频道喵！`)
		fallthrough
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

// GetImageURL 获取事件上下文中的图片链接
func GetImageURL(ctx *zero.Ctx) []string {
	return GetSth[[]string](ctx, `image_url`)
}

// 检查接口中是否存在错误，如果有则记录至日志
func checkErr(v []any) {
	for _, i := range v {
		if err, ok := i.(error); ok {
			Error(err)
		}
	}
}

// ScanQRCode 扫描二维码
func ScanQRCode(name string) (fmt.Stringer, error) {
	var (
		msg = message.Image(name)
		n   = core.FilePath(`data`, `zbp`, `code.png`)
	)
	if err := core.GetImage(msg.Data[`file`], n); nil != err {
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
