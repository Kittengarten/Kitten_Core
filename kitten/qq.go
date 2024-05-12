package kitten

import (
	"slices"
	"strconv"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/RomiChan/syncx"
	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
)

type (
	QQ int64 // QQ 是一个表示 QQ 的 int64

	qqInfo struct {
		Info gjson.Result // 信息
		T    time.Time    // 上次更新时间
	} // QQ 信息

	groupList struct {
		List []gjson.Result // 每个群员的信息
		T    time.Time      // 上次更新时间
	} // 群成员列表
)

const (
	LastSentTime = `last_sent_time` // 上次发送时间
	UserID       = `user_id`        // QQ
	expire       = time.Hour        // 缓存过期时间
)

var (
	StrangerInfo    syncx.Map[QQ, qqInfo]       // 各陌生人信息缓存
	GroupMemberList syncx.Map[int64, groupList] // 各群的成员列表缓存
)

// Set 设置 QQ 的 int64 类型原始值
func (u *QQ) Set(v int64) {
	*u = QQ(v)
}

// Int 获取 QQ 的 int64 类型表示
func (u *QQ) Int() int64 {
	return int64(*u)
}

// Int 获取 QQ 的 string 类型表示
func (u *QQ) String() string {
	return strconv.FormatInt(u.Int(), 10)
}

// （私有）获取陌生人信息
func (u *QQ) info(ctx *zero.Ctx) qqInfo {
	if !CheckCtx(ctx, Caller) {
		// 没有 APICaller ，无法获取
		return qqInfo{}
	}
	// 从缓存获取该 QQ 的信息
	si, ok := StrangerInfo.Load(*u)
	if !ok {
		// 如果获取不到，同步更新
		u.updateInfo(ctx)
		si, _ = StrangerInfo.Load(*u)
	}
	// 如果缓存已经过期，异步更新缓存的陌生人信息
	if expire < time.Since(si.T) {
		go u.updateInfo(ctx)
	}
	return si
}

// （私有）更新陌生人信息
func (u *QQ) updateInfo(ctx *zero.Ctx) {
	StrangerInfo.Store(*u,
		qqInfo{
			Info: ctx.GetStrangerInfo(u.Int(), true),
			T:    time.Now(),
		},
	)
}

// （私有）获取群成员信息
func (u *QQ) memberInfo(ctx *zero.Ctx) qqInfo {
	list := MemberList(ctx)
	if 0 == len(list.List) {
		// 如果本群成员列表为空，退化至陌生人
		return u.info(ctx)
	}
	// 从本群成员列表中查找
	i := slices.IndexFunc(list.List, func(i gjson.Result) bool {
		return i.Get(UserID).Int() == u.Int()
	})
	if -1 == i {
		// 如果本群成员列表中找不到，退化至陌生人
		return u.info(ctx)
	}
	return qqInfo{
		Info: list.List[i],
		T:    list.T,
	}
}

// Age 获取年龄
func (u *QQ) Age(ctx *zero.Ctx) int64 {
	return u.info(ctx).Info.Get(`age`).Int()
}

// IsAdult 是成年人
func (u *QQ) IsAdult(ctx *zero.Ctx) bool {
	return 18 <= u.Age(ctx)
}

// IsFemale 是女性
func (u *QQ) IsFemale(ctx *zero.Ctx) bool {
	return `female` == u.info(ctx).Info.Get(`sex`).String()
}

// IsLoli 是萝莉
func (u *QQ) IsLoli(ctx *zero.Ctx) bool {
	return u.IsFemale(ctx) && 0 < u.Age(ctx) && 18 > u.Age(ctx)
}

// TitleCardOrNickName 从 QQ 获取【头衔】群昵称 | 昵称
func (u *QQ) TitleCardOrNickName(ctx *zero.Ctx) string {
	if !CheckCtx(ctx, Caller) {
		// 没有 APICaller ，无法获取
		return ``
	}
	// 修剪后的昵称
	name := core.CleanAll(u.info(ctx).Info.Get(`nickname`).Str, false)
	if 0 >= ctx.Event.GroupID {
		// 不是群聊，直接返回昵称
		return name
	}
	// 是群聊，获取该 QQ 在群内的信息
	var (
		gmi   = u.memberInfo(ctx)
		title = gmi.Info.Get(`title`).Str // 头衔
	)
	if `` != title {
		// 如果头衔存在，则添加实心方头括号
		title = `【` + title + `】	`
	}
	// 获取修剪后的群昵称
	if card := core.CleanAll(gmi.Info.Get(`card`).Str, false); `` != card {
		// 如果不为空，返回【头衔】	群昵称
		return title + card
	}
	// 返回【头衔】	昵称
	return title + name
}

// MemberList 获取群成员列表
func MemberList(ctx *zero.Ctx) groupList {
	if !CheckCtx(ctx, Caller) {
		// 没有 APICaller ，无法获取
		return groupList{}
	}
	// 从缓存获取该群成员列表
	gmi, ok := GroupMemberList.Load(ctx.Event.GroupID)
	if !ok {
		// 如果获取不到，同步更新
		updateMemberList(ctx)
		gmi, _ = GroupMemberList.Load(ctx.Event.GroupID)
	}
	// 如果缓存已经过期，异步更新缓存的该群成员列表
	if expire < time.Since(gmi.T) {
		go updateMemberList(ctx)
	}
	return gmi
}

// （私有）更新群成员列表
func updateMemberList(ctx *zero.Ctx) {
	GroupMemberList.Store(ctx.Event.GroupID,
		groupList{
			List: ctx.GetThisGroupMemberListNoCache().Array(),
			T:    time.Now(),
		},
	)
}
