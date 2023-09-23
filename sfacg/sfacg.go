// Package sfacg SF 轻小说更新播报、小说信息查询、小说更新查询
package sfacg

import (
	"cmp"
	"fmt"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	mapset "github.com/deckarep/golang-set"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	replyServiceName = `sfacg` // 插件名
	brief            = `SF 轻小说报更`
	configFile       = `config.yaml` // 配置文件名
	ag               = `参数`
	cNovel           = `小说`
	cUpdateTest      = `更新测试`
	cUpdatePreview   = `更新预览`
	cAddUpadte       = `添加报更`
	cCancelUpadte    = `取消报更`
	cQueryUpadte     = `查询报更`
	without          = `这里没有添加小说报更喵～`
)

var (
	config = kitten.GetMainConfig() // 主配置
	cu     chan books               // 报更更新的信号
)

func init() {
	var (
		cpf  = config.CommandPrefix
		help = fmt.Sprintf(`%s%s [%s]，可获取信息
%s%s [%s] // 可测试报更功能
%s%s [%s] // 可预览更新内容
%s%s // 可查询当前小说自动报更
————
管理员可用：
%s%s [%s] // 可添加小说自动报更
%s%s [%s] // 可取消小说自动报更`,
			cpf, cNovel, ag,
			cpf, cUpdateTest, ag,
			cpf, cUpdatePreview, ag,
			cpf, cQueryUpadte,
			cpf, cAddUpadte, ag,
			cpf, cCancelUpadte, ag)
		// 注册插件
		engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  false,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: replyServiceName,
		}).ApplySingle(ctxext.DefaultSingle)
		lbg = ctxext.NewLimiterManager(time.Minute, 1).LimitByGroup // 群内共通限速器
	)

	go track(engine)

	// 测试小说报更功能
	engine.OnCommand(cUpdateTest).SetBlock(true).Limit(lbg).Handle(func(ctx *zero.Ctx) {
		nv, err := getNovel(ctx)
		if nil != err {
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		report, _ := nv.update()
		kitten.SendMessage(ctx, true, message.Image(nv.coverURL), message.Image(nv.headURL), message.Text(report))
	})

	// 预览小说更新功能
	engine.OnCommand(cUpdatePreview).SetBlock(true).Limit(lbg).Handle(func(ctx *zero.Ctx) {
		n, err := getNovel(ctx)
		if nil != err {
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		if r := n.preview; `` != r {
			kitten.SendText(ctx, true, r)
			return
		}
		kitten.SendWithImageFail(ctx, `不存在的喵！`)
	})

	// 小说信息功能
	engine.OnCommand(cNovel).SetBlock(true).Limit(lbg).Handle(func(ctx *zero.Ctx) {
		n, err := getNovel(ctx)
		if nil != err {
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		kitten.SendMessage(ctx, true, message.Image(n.coverURL), message.Text(n.information()))
	})

	// 设置报更
	engine.OnCommand(cAddUpadte).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var o int64 // 发送对象
		if o = getO(ctx); slices.Contains([]int64{0, 1}, o) {
			return
		}
		c, err := loadConfig(kitten.FilePath(engine.DataFolder(), configFile)) // 报更配置
		if nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		var (
			lc       = len(c)
			b        = make([]string, lc, lc)          // 书号切片
			groupSet = make(map[string]mapset.Set, lc) // 书号:群号集合
		)
		for k := range c {
			// 本书书号
			b[k] = c[k].BookID
			// 本书报更的群号集合
			groupSet[c[k].BookID] = mapset.NewSet(c[k].GroupID)
		}
		novel, err := getNovel(ctx) // 小说实例
		if nil != err {
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		if mapset.NewSet(b).Contains(novel.id) {
			// 已经有该小说
			if !groupSet[novel.id].Add(o) {
				// 已有该群号，添加失败
				kitten.SendWithImageFail(ctx, `《%s》已经添加报更了喵！`, novel.name)
				return
			}
			// 尚无该群号，添加成功
			for k := range c {
				var (
					// 获取本书的报更对象接口切片
					gi  = groupSet[c[k].BookID].ToSlice()
					lgi = len(gi)
					gs  = make([]int64, lgi, lgi)
				)
				// 将接口切片强转为 int64 群号切片
				for i := range gi {
					gs[i] = gi[i].(int64)
				}
				// 群号排序
				slices.Sort(gs)
				// 回写
				c[k].GroupID = gs
			}
		} else {
			// 没有该小说，新建并添加
			c = append(c, book{
				BookID:   novel.id,
				BookName: novel.name,
				GroupID:  []int64{o},
			})
		}
		c.saveConfig(ctx, `添加`, novel, engine)
	})

	// 移除报更
	engine.OnCommand(cCancelUpadte).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var o int64 // 发送对象
		if o = getO(ctx); slices.Contains([]int64{0, 1}, o) {
			// 获取发送对象，如果当前发送对象不允许发送，则直接返回
			return
		}
		c, err := loadConfig(kitten.FilePath(engine.DataFolder(), configFile)) // 报更配置
		if nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		lc := len(c)
		if 0 >= lc {
			ctx.Send(without)
			return
		}
		novel, err := getNovel(ctx) // 小说实例
		if nil != err {
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		for k := range c {
			if novel.id != c[k].BookID {
				// 如果当前遍历的书不是取消的书，则直接判断下一本
				continue
			}
			groupSet := make(map[string]mapset.Set, lc) // 书号:群号集合
			// 如果当前遍历的书应该取消报更，构建群号集合
			groupSet[c[k].BookID] = mapset.NewSet(c[k].GroupID)
			// 移除前，群的数量
			n := groupSet[c[k].BookID].Cardinality()
			// 移除在当前发送对象的报更
			groupSet[c[k].BookID].Remove(o)
			if 0 < groupSet[c[k].BookID].Cardinality() {
				// 如果移除后，群的数量非 0（还有剩余的报更对象）
				if n == groupSet[c[k].BookID].Cardinality() {
					// 并没有移除成功，当前发送对象本来就没有在报更
					kitten.SendText(ctx, false, `未追更本书喵！`)
					return
				}
				var (
					// 移除成功，获取本书的报更对象接口切片
					gi  = groupSet[c[k].BookID].ToSlice()
					lgi = len(gi)
					gs  = make([]int64, lgi, lgi)
				)
				// 将接口切片强转为 int64 群号切片
				for i := range gi {
					gs[i] = gi[i].(int64)
				}
				// 群号排序
				slices.Sort(gs)
				// 回写
				c[k].GroupID = gs
			} else {
				// 如果移除后，不再有报更对象（也可能本来就没有报更对象），则移除该小说
				c = slices.Delete(c, k, 1+k)
			}
		}
		c.saveConfig(ctx, `取消`, novel, engine)
	})

	// 查询报更
	engine.OnCommand(cQueryUpadte).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var o int64 // 发送对象
		if o = getO(ctx); slices.Contains([]int64{0, 1}, o) {
			return
		}
		c, err := loadConfig(kitten.FilePath(engine.DataFolder(), configFile)) // 报更配置
		if nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		if 0 >= len(c) {
			ctx.Send(without)
			return
		}
		const h = `【报更列表】`
		var r strings.Builder
		r.WriteString(h)
		for k := range c {
			var t string
			if `` == c[k].UpdateTime {
				t = `未知`
			} else {
				t = c[k].UpdateTime
			}
			for i := range c[k].GroupID {
				if o != c[k].GroupID[i] {
					continue
				}
				r.WriteString(fmt.Sprintf(`《%s》，书号 %s`, c[k].BookName, c[k].BookID))
				r.WriteString(fmt.Sprint(`上次更新：`, t))
			}
		}
		ctx.Send(r.String())
	})
}

/*
获取小说

如果传入值不为书号，则先获取书号
*/
func getNovel(ctx *zero.Ctx) (nv novel, err error) {
	a, ok := ctx.State[`args`]
	if !ok {
		return novel{}, fmt.Errorf(`获取小说时，上下文状态 %v 不包含参数`, ctx.State)
	}
	as, ok := a.(string)
	if !ok {
		return novel{}, fmt.Errorf(`获取小说时，参数 %v 无法断言为字符串`, a)
	}
	if _, err = strconv.Atoi(as); nil != err {
		zap.S().Debugf(`获取小说时，参数字符串 %s 无法转换为书号，尝试作为搜索关键词`, as)
		if as, err = keyWord(as).findBookID(); nil != err {
			return novel{}, err
		}
	}
	nv = *novelPool.Get().(*novel)
	if err := nv.init(as); nil != err {
		return novel{}, err
	}
	defer novelPool.Put(&nv)
	return
}

// 报更
func track(e *control.Engine) {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); nil != err {
			zap.S().Error(replyServiceName, `报更协程出现错误喵！`, err)
			debug.PrintStack()
		}
	}()
	if err := func() error {
		p := kitten.FilePath(e.DataFolder(), configFile)
		return kitten.InitFile(&p, kitten.Empty)
	}; nil != err() {
		zap.S().Error(err())
		return
	}
	data, err := loadConfig(kitten.FilePath(e.DataFolder(), configFile))
	if nil != err {
		zap.S().Error(err)
		return
	}
	fmt.Println(fmt.Sprintf(`======================[%s]======================
* OneBot + ZeroBot + Go
一共有 %d 本小说
=======================================================`,
		config.NickName[0],
		len(data)))
	process.GlobalInitMutex.Lock()
	bot := zero.GetBot(config.SelfID)
	zap.S().Debug(`获取的 Bot 实例：`, bot)
	process.GlobalInitMutex.Unlock()
	t := time.NewTicker(5 * time.Second) // 每 5 秒检测一次
	// 报更
	for {
		select {
		case data = <-cu: // 接收到更新配置则使用
		case <-t.C: // 接收到定时器信号则释放
		}
		for i := range data {
			// 从小说池初始化小说
			nv := *novelPool.Get().(*novel)
			if err = nv.init(data[i].BookID); nil != err {
				zap.S().Error(err)
				continue
			}
			// 更新判定，并防止误报
			if nv.newChapter.url == data[i].RecordURL {
				continue
			}
			report, d := nv.update()
			// 距上次更新时间小于等于 1 秒则不报更，防止异常信息发送
			if time.Second > d {
				continue
			}
			// 群号排序
			slices.Sort(data[i].GroupID)
			for k := range data[i].GroupID {
				if 0 < data[i].GroupID[k] {
					bot.SendGroupMessage(data[i].GroupID[k], message.Message{
						message.Image(nv.coverURL), message.Image(nv.headURL), message.Text(report)})
				} else {
					bot.SendPrivateMessage(-data[i].GroupID[k], message.Message{
						message.Image(nv.coverURL), message.Image(nv.headURL), message.Text(report)})
				}
			}
			data[i].BookName = nv.name
			data[i].RecordURL = nv.newChapter.url
			data[i].UpdateTime = nv.newChapter.update.Format(kitten.Layout)
			// 按更新时间倒序排列
			slices.SortFunc(data, func(j, i book) int { return cmp.Compare(i.UpdateTime, j.UpdateTime) })
			updateConfig, err := yaml.Marshal(data)
			if nil == err {
				zap.S().Infof(`记录 %s 成功喵！`, e.DataFolder()+configFile)
			} else {
				zap.S().Warnf(`记录 %s 失败喵！`, e.DataFolder()+configFile)
				continue
			}
			kitten.FilePath(e.DataFolder(), configFile).Write(updateConfig)
		}
	}
}

/*
获取发送对象

返回正整数代表群，返回负整数代表私聊

返回默认值 0 代表不支持的对象（目前是频道）

返回 1 代表在群中无权限
*/
func getO(ctx *zero.Ctx) int64 {
	switch ctx.Event.DetailType {
	case "private":
		return -ctx.Event.UserID
	case "group":
		if !zero.AdminPermission(ctx) {
			kitten.SendWithImageFail(ctx, `你没有管理员权限喵！`)
			return 1
		}
		return ctx.Event.GroupID
	case "guild":
		ctx.Send(kitten.Guild)
		return 0
	default:
		return 0
	}
}
