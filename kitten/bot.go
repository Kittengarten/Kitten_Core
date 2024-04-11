package kitten

import (
	"fmt"
	"os/exec"
	"reflect"
	"strconv"

	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	noEvent      = `非消息的上下文中获取的 bot 实例无 *Event，不可使用`
	Caller  Item = iota // APICaller
	Event               // *Event
)

// Restart 重启 systemd 服务
func Restart(s string) {
	output, err := exec.Command(`sudo`, `systemctl`, `restart`, s).Output()
	if nil != err {
		Error(string(output))
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
	if !CheckCtx(ctx, Event) || !CheckCtx(ctx, Caller) {
		// 没有事件或 APICaller ，无法发送
		Info(text...)
		return message.NewMessageIDFromInteger(0)
	}
	switch atUser := message.At(ctx.Event.UserID); ctx.Event.DetailType {
	case `private`:
		return ctx.Send(Text(text...))
	case `group`, `guild`:
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
		Infof(format, a...)
		return message.NewMessageIDFromInteger(0)
	}
	switch atUser := message.At(ctx.Event.UserID); ctx.Event.DetailType {
	case `private`:
		return ctx.Send(TextOf(format, a...))
	case `group`, `guild`:
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
	case `private`:
		return ctx.Send(m)
	case `group`, `guild`:
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
	img, err := imagePath.Image(`no.png`)
	if nil != err {
		return SendText(ctx, true, err)
	}
	return SendMessage(ctx, true, img, Text(text...))
}

/*
SendWithImageFailOf 发送带有失败图片的格式化文字消息

非消息的事件中获取的 bot 实例可能无效
*/
func SendWithImageFailOf(ctx *zero.Ctx, format string, a ...any) message.MessageID {
	img, err := imagePath.Image(`no.png`)
	if nil != err {
		return SendText(ctx, true, err)
	}
	return SendMessage(ctx, true, img, TextOf(format, a...))
}

/*
DoNotKnow 喵喵不知道哦

非消息的事件中获取的 bot 实例可能无效
*/
func DoNotKnow(ctx *zero.Ctx) message.MessageID {
	img, err := imagePath.Image(`哈——？.png`)
	if nil != err {
		return SendText(ctx, true, err)
	}
	return SendMessage(ctx, true, img, TextOf(`%s不知道哦`, zero.BotConfig.NickName[0]))
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
	case "private":
		return -ctx.Event.UserID
	case "group":
		if !admin || zero.AdminPermission(ctx) {
			return ctx.Event.GroupID
		}
		SendWithImageFail(ctx, `管理员才能使用喵！`)
		return 1
	case "guild":
		SendWithImageFail(ctx, `暂不支持频道喵！`)
		fallthrough
	default:
		return 0
	}
}

// TitleCardOrNickName 从 QQ 获取【头衔】群昵称 | 昵称
func (u QQ) TitleCardOrNickName(ctx *zero.Ctx) string {
	if !CheckCtx(ctx, Caller) {
		// 没有 APICaller ，无法获取
		return ``
	}
	// 修剪后的昵称
	name := core.CleanAll(ctx.GetStrangerInfo(u.Int(), true).Get(`nickname`).Str, false)
	if 0 >= ctx.Event.GroupID {
		// 不是群聊，直接返回昵称
		return name
	}
	// 是群聊，获取该 QQ 在群内的资料
	var (
		gmi   = ctx.GetThisGroupMemberInfo(u.Int(), true) // 本群成员信息
		title = gmi.Get(`title`).Str                      // 头衔
	)
	if `` != title {
		// 如果头衔存在，则添加实心方头括号
		title = `【` + title + `】	`
	}
	// 获取修剪后的群昵称
	if card := core.CleanAll(gmi.Get(`card`).Str, false); `` != card {
		// 如果不为空，返回【头衔】	群昵称
		return title + card
	}
	// 返回【头衔】	昵称
	return title + name
}

// CheckCtx 检查事件的某个项目是否有效且不为空
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

// Int 获取 QQ 的 int64 类型表示
func (u QQ) Int() int64 {
	return int64(u)
}

// Int 获取 QQ 的 string 类型表示
func (u QQ) String() string {
	return strconv.FormatInt(u.Int(), 10)
}

// （私有）获取信息
func (u QQ) info(ctx *zero.Ctx) gjson.Result {
	if !CheckCtx(ctx, Caller) {
		// 没有 APICaller ，无法获取
		return gjson.Result{}
	}
	return ctx.GetStrangerInfo(u.Int(), true)
}

// IsAdult 是成年人
func (u QQ) IsAdult(ctx *zero.Ctx) bool {
	return 18 <= u.info(ctx).Get(`age`).Int()
}

// IsFemale 是女性
func (u QQ) IsFemale(ctx *zero.Ctx) bool {
	return `female` == u.info(ctx).Get(`sex`).String()
}

// IsLoli 是萝莉
func (u QQ) IsLoli(ctx *zero.Ctx) bool {
	return u.IsFemale(ctx) &&
		0 < u.info(ctx).Get(`age`).Int() &&
		18 > u.info(ctx).Get(`age`).Int()
}
