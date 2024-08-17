// Package track 小说更新播报、小说信息查询、小说更新查询
package track

import (
	"fmt"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	replyServiceName = `track` // 插件名
	brief            = `小说报更`
	configFile       = `config.yaml` // 配置文件名
	pf               = `[平台]`
	ag               = `[关键词|书号]`
	cNovel           = `小说`
	cUpdateTest      = `更新测试`
	cUpdatePreview   = `更新预览`
	cAddUpdate       = `添加报更`
	cCancelUpdate    = `取消报更`
	cQueryUpdate     = `查询报更`
	without          = `这里没有添加小说报更喵～`
	errConfig        = `报更配置文件错误喵！`
	errLoad          = `加载` + errConfig
	errSave          = `保存` + errConfig
)

var (
	// 指令前缀
	p = kitten.MainConfig().CommandPrefix
	// 帮助
	help = p + cNovel + ` ` + pf + ` ` + ag + ` // 可获取信息
` + p + cUpdateTest + ` ` + pf + ` ` + ag + ` // 可测试报更功能
` + p + cUpdatePreview + ` ` + pf + ` ` + ag + ` // 可预览更新内容
` + p + cQueryUpdate + ` // 可查询当前小说自动报更
————
管理员或私聊可用：
` + p + cAddUpdate + ` ` + pf + ` ` + ag + ` // 可添加小说自动报更
` + p + cCancelUpdate + ` ` + pf + ` ` + ag + ` // 可取消小说自动报更`
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             brief,
		Help:              help,
		PrivateDataFolder: replyServiceName,
	}).ApplySingle(ctxext.DefaultSingle)
	// 配置文件路径
	configPath = core.FilePath(engine.DataFolder(), configFile)
	// 报更更新的信号
	cu = make(chan books)
	// 读写锁
	mu sync.RWMutex
)

func init() {
	go track()

	// 更新测试
	engine.OnCommand(cUpdateTest).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.User)).
		Limit(kitten.GetLimiter(kitten.GroupNormal)).
		Handle(updateTest)

	// 更新预览
	engine.OnCommand(cUpdatePreview).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.User)).
		Limit(kitten.GetLimiter(kitten.GroupNormal)).
		Handle(updatePreview)

	// 小说信息功能
	engine.OnCommand(cNovel).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.User)).
		Limit(kitten.GetLimiter(kitten.GroupNormal)).
		Handle(novelInfo)

	// 添加报更
	engine.OnCommand(cAddUpdate, zero.UserOrGrpAdmin).SetBlock(true).
		Limit(ctxext.LimitByGroup).Handle(add)

	// 取消报更
	engine.OnCommand(cCancelUpdate, zero.UserOrGrpAdmin).SetBlock(true).
		Limit(ctxext.LimitByGroup).Handle(cancel)

	// 查询报更
	engine.OnCommand(cQueryUpdate).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.GroupSlow)).Handle(query)
}

// 更新测试
func updateTest(ctx *zero.Ctx) {
	nv, err := getNovel(ctx)
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	kitten.SendMessage(ctx, true,
		kitten.Image(nv.coverURL),
		kitten.Image(nv.headURL),
		kitten.Text(nv.update()))
}

// 更新预览
func updatePreview(ctx *zero.Ctx) {
	n, err := getNovel(ctx)
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	if r := n.preview; `` != r {
		kitten.SendTextOf(ctx, true, `《%s》
%s
%s`,
			n.name,
			&n.newChapter,
			r,
		)
		return
	}
	kitten.SendWithImageFail(ctx, `不存在的喵！`)
}

// 小说信息
func novelInfo(ctx *zero.Ctx) {
	n, err := getNovel(ctx)
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	kitten.SendMessage(ctx, true, kitten.Image(n.coverURL), kitten.Text(&n))
}

// 添加报更
func add(ctx *zero.Ctx) {
	o := kitten.GetObject(ctx) // 发送对象
	mu.Lock()
	defer mu.Unlock()
	c, err := core.Load[books](configPath, core.Empty) // 报更配置
	if nil != err {
		kitten.SendWithImageFail(ctx, errLoad, err)
		return
	}
	nv, err := getNovel(ctx) // 小说实例
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	if i := slices.IndexFunc(c, func(b book) bool {
		return b.Platform == nv.platform && b.BookID == nv.id
	}); -1 == i {
		// 没有该小说，新建并添加
		c = append(c, book{
			Platform: nv.platform,
			BookID:   nv.id,
			BookName: nv.name,
			Writer:   nv.writer,
			Users:    []kitten.QQ{o},
		})
	} else {
		// 已经有该小说
		if slices.Contains(c[i].Users, o) {
			// 已有该用户，无需添加
			kitten.SendWithImageFailOf(ctx, `《`+nv.name+`》已经添加报更了喵！`)
			return
		}
		// 尚无该用户，需要添加
		c[i].Users = append(c[i].Users, o)
		slices.Sort(c[i].Users)
	}
	if err := c.saveConfig(); nil != err {
		kitten.SendWithImageFail(ctx, `添加《`+nv.name+`》时`+errSave, err)
		return
	}
	kitten.SendText(ctx, false, `添加《`+nv.name+`》报更成功喵！`)
}

// 取消报更
func cancel(ctx *zero.Ctx) {
	o := kitten.GetObject(ctx) // 发送对象
	mu.Lock()
	defer mu.Unlock()
	c, err := core.Load[books](configPath, core.Empty) // 报更配置
	if nil != err {
		kitten.SendWithImageFail(ctx, errLoad, err)
		return
	}
	if 0 == len(c) {
		kitten.SendText(ctx, false, without)
		return
	}
	nv, err := getNovel(ctx) // 小说实例
	if nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	// 本书下标
	i := slices.IndexFunc(c, func(b book) bool {
		return b.Platform == nv.platform && b.BookID == nv.id
	})
	if -1 == i {
		kitten.SendWithImageFailOf(ctx, `未在追更《`+nv.name+`》喵！`)
		return
	}
	// 用户下标
	uid := slices.Index(c[i].Users, o)
	if -1 == uid {
		kitten.SendWithImageFailOf(ctx, `未在追更《`+nv.name+`》喵！`)
		return
	}
	// 移除在当前发送对象的报更
	if 0 < len(slices.Delete(c[i].Users, uid, 1+uid)) {
		// 用户排序
		slices.Sort(c[i].Users)
	} else {
		// 如果移除后，不再有报更对象（也可能本来就没有报更对象），则整体移除该小说
		c = slices.Delete(c, i, 1+i)
	}
	if err := c.saveConfig(); nil != err {
		kitten.SendWithImageFail(ctx, `取消《`+nv.name+`》时`+errSave, err)
		return
	}
	kitten.SendText(ctx, false, `取消《`+nv.name+`》报更成功喵！`)
}

// 查询报更
func query(ctx *zero.Ctx) {
	o := kitten.GetObject(ctx) // 发送对象
	mu.RLock()
	c, err := core.Load[books](configPath, core.Empty) // 报更配置
	mu.RUnlock()
	if nil != err {
		kitten.SendWithImageFail(ctx, errLoad, err)
		return
	}
	if 0 == len(c) {
		kitten.SendText(ctx, false, without)
		return
	}
	const h = `【报更列表】`
	var r strings.Builder
	r.Grow(64 * len(c))
	r.WriteString(h)
	for _, b := range c {
		if !slices.Contains(b.Users, o) {
			// 如果本书不在这里报更，则直接遍历至下一本书
			continue
		}
		r.WriteByte('\n')
		r.WriteString(b.String())
	}
	kitten.SendText(ctx, true, &r)
}

/*
获取小说

如果传入值不为书号，则先获取书号
*/
func getNovel(ctx *zero.Ctx) (novel, error) {
	args := kitten.GetArgsSlice(ctx)
	if 2 != len(args) {
		return novel{}, fmt.Errorf(`本命令参数数量：2
%s %s
传入的参数数量：%d
参数数量错误喵！`,
			pf, ag,
			len(args))
	}
	p := func() platform {
		switch {
		case strings.ContainsAny(args[0], `菠萝包`),
			strings.Contains(strings.ToUpper(args[0]), `SF`),
			strings.Contains(strings.ToLower(args[0]), `blb`):
			return sf
		case strings.Contains(args[0], `刺猬`),
			strings.ContainsAny(args[0], `猫客`),
			strings.Contains(strings.ToUpper(args[0]), `CWM`),
			strings.Contains(strings.ToLower(args[0]), `ciweimao`):
			return cwm
		default:
			return platform(args[0])
		}
	}()
	if _, err := strconv.Atoi(args[1]); nil != err {
		kitten.Debugf(`获取小说时，参数字符串 %s 无法转换为书号，尝试作为搜索关键词`, args[1])
		if args[1], err = keyword(args[1]).findBookID(p); nil != err {
			return novel{}, err
		}
	}
	nv := *novelPool.Get().(*novel)
	defer novelPool.Put(&nv)
	return nv, nv.init(p, args[1])
}

// 报更
func track() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); nil != err {
			kitten.Error(replyServiceName, ` 协程出现错误喵！`, err, string(debug.Stack()))
		}
	}()
	// 初始化报更配置文件
	if err := core.InitFile(&configPath, core.Empty); nil != err {
		kitten.Error(`初始化报更配置文件时发生错误喵！`, err)
		return
	}
	mu.RLock()
	data, err := core.Load[books](configPath, core.Empty)
	mu.RUnlock()
	if nil != err {
		kitten.Error(errLoad, err)
		return
	}
	fmt.Printf(`======================[%s]======================
* OneBot + ZeroBot + Go
一共有 %d 本小说
=======================================================
`,
		kitten.MainConfig().NickName[0],
		len(data))
	process.GlobalInitMutex.Lock()
	process.GlobalInitMutex.Unlock()
	var (
		t   = time.NewTicker(5 * time.Second) // 每 5 秒检测一次
		u   = kitten.MainConfig().SelfID
		bot = zero.GetBot(u.Int())
	)
	kitten.Debugln(`获取的 Bot 实例：`, bot)
	// 报更
	for {
		select {
		case data = <-cu: // 接收到更新配置则使用
		case <-t.C: // 接收到定时器信号则释放
		}
		data.report(bot) // 执行报更
	}
}

// 执行报更
func (c *books) report(ctx *zero.Ctx) {
	for i, b := range *c {
		// 从小说池初始化小说
		var (
			nv  = *novelPool.Get().(*novel)
			err = nv.init(platform(b.Platform), b.BookID)
		)
		if nil != err {
			kitten.Error(err)
			continue
		}
		// 更新判定
		if nv.newChapter.url == b.RecordURL {
			continue
		}
		// 用户排序
		slices.Sort(b.Users)
		// 消息构造
		msg := message.Message{
			kitten.Image(nv.coverURL),
			kitten.Image(nv.headURL),
			kitten.Text(nv.update()),
		}
		// 距上次更新时间小于等于 1 秒则不报更，防止异常信息发送
		if time.Second > nv.timeGap {
			continue
		}
		for _, id := range b.Users {
			core.RandomDelay(time.Second)
			id.SendMessage(ctx, msg)
		}
		// 写入小说更新数据
		(*c)[i].BookName = nv.name
		(*c)[i].Writer = nv.writer
		(*c)[i].RecordURL = nv.newChapter.url
		(*c)[i].UpdateTime = nv.newChapter.update
		// 将小说重新收回小说池
		novelPool.Put(&nv)
		// 按更新时间倒序排列
		c.SortByUpdate()
		// 异步保存配置
		go func() {
			err = c.saveConfig()
		}()
		if nil != err {
			kitten.Error(errSave, err)
			continue
		}
		kitten.Infof(`更新《` + nv.name + `》成功喵！`)
	}
}
