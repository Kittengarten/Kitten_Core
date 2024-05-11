// repeat 喵类的本质是复读姬
package repeat

import (
	"fmt"
	"html"
	"math/rand/v2"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	replyServiceName = `repeat`      // 插件名
	configFile       = `config.yaml` // 配置文件名
	cRepeat          = `复读`
	brief            = `喵类的本质是复读姬`
	MaxTimes         = 10 // 设置触发复读的最多次数限制
)

type (
	// 消息统计
	stat struct {
		t   uint            // 次数
		msg message.Message // 消息
	}

	// 复读姬配置
	config struct {
		Times  uint    `yaml:"times"`  // 触发复读的次数
		Chance float64 `yaml:"chance"` // 触发复读的概率
	}
)

var (
	// 帮助
	help = kitten.MainConfig().CommandPrefix + cRepeat + ` [次数] [概率]`
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             brief,
		Help:              help,
		PrivateDataFolder: replyServiceName,
	})
	// 触发复读的次数
	times uint = 2
	// 触发复读的概率
	chance = 0.5
	// 消息缓存
	m sync.Map
	// 互斥锁（只写不读）
	mu sync.Mutex
)

func init() {
	repeatConfig, err := load(configFile) // 复读姬配置文件
	if nil == err {
		times, chance = repeatConfig.Times, repeatConfig.Chance
	} else {
		kitten.Error(`加载复读姬配置文件错误喵！`, err)
	}

	// 复读设置
	engine.OnCommand(cRepeat, zero.SuperUserPermission).SetBlock(true).
		Limit(kitten.GetLimiter(kitten.GroupSlow)).Handle(repeatSet)

	// 复读
	engine.OnMessage(zero.OnlyGroup).SetBlock(false).Handle(repeat)
}

// 复读设置
func repeatSet(ctx *zero.Ctx) {
	args := slices.DeleteFunc(strings.Split(kitten.GetArgs(ctx), ` `),
		func(s string) bool {
			return `` == s
		})
	if 2 != len(args) {
		kitten.SendWithImageFailOf(ctx, `本命令参数数量：%d
传入的参数数量：%d`,
			2,
			len(args),
		)
		return
	}
	t, err := strconv.ParseUint(args[0], 10, core.PlatformBits)
	if nil != err {
		kitten.SendWithImageFailOf(ctx, `[次数] 错误：
%v
%s`,
			err,
			help,
		)
		return
	}
	if 2 > t || MaxTimes < t {
		kitten.SendWithImageFailOf(ctx, `[次数] 错误：最少为 2，最多为 %d 喵！`, MaxTimes)
		return
	}
	times = uint(t)
	if chance, err = strconv.ParseFloat(args[1], 64); nil != err {
		kitten.SendWithImageFailOf(ctx, `[概率] 错误：
%v
%s`,
			err,
			help,
		)
		return
	}
	if 0 > chance {
		chance = 0
		kitten.SendWithImageFailOf(ctx, `[概率] 警告：不能 ＜ 0 喵！`)
		return
	}
	if 1 < chance {
		chance = 1
		kitten.SendWithImageFailOf(ctx, `[概率] 警告：不能 ＞ 1 喵！`)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	err = config{
		Times:  times,
		Chance: chance,
	}.save(ctx)
	if nil != err {
		kitten.SendWithImageFail(ctx, `I/O 错误：`, err)
		return
	}
	kitten.SendTextOf(ctx, true, `%s将会开始以 %.2f%% 概率复读重复 %d 次的消息喵！`,
		zero.BotConfig.NickName[0], 100*chance, times)
}

func repeat(ctx *zero.Ctx) {
	var (
		g     = ctx.Event.GroupID // 群号
		c, ok = m.Load(g)         // 尝试获取本群的缓存
		s     = stat{
			t:   1,
			msg: ctx.Event.Message,
		} // 缓存的消息统计
	)
	if ok && compare(c.(stat).msg, ctx.Event.Message) {
		// 如果消息与缓存的内容一致，增加一次复读计数
		s = c.(stat)
		s.t++
	}
	// 更新缓存
	m.Store(g, s)
	if times > s.t || chance <= rand.Float64() {
		// 如果没有达到复读阈值，或者没有按概率触发复读，则返回
		return
	}
	if 1 == len(s.msg) && `image` == s.msg[0].Type {
		// 如果是单张图片，重新构造图片消息段并发送
		ctx.Send(message.Image(html.UnescapeString(core.MidText(`url=`, `]`, s.msg[0].String()))))
		return
	}
	ctx.Send(s.msg)
}

// 比较两个消息段切片是否相等
func compare(x, y message.Message) bool {
	if len(x) != len(y) {
		// 如果两个消息段切片的长度不同，则不相等
		return false
	}
	for i, t := range x {
		tx, ty := x[i].Type, y[i].Type
		if tx != ty {
			// 如果两个消息段类型不同，则不相等
			return false
		}
		switch t.Type {
		// 按类型的特殊比较路径
		case `image`:
			if core.MidText(`file=`, `.image`, x[i].String()) !=
				core.MidText(`file=`, `.image`, y[i].String()) {
				// 如果图片 MD5 不同，则不相等
				return false
			}
		case `record`, `video`, `anonymous`, `share`, `contact`,
			`location`, `music`, `forward`, `node`, `xml`, `json`:
			// 不复读的类型
			return false
		default:
			if !reflect.DeepEqual(x[i].Data, y[i].Data) {
				return false
			}
		}
	}
	return true
}

// 加载复读姬配置
func load(configFile string) (c config, err error) {
	d, err := getPath(configFile).Read()
	if nil != err {
		return
	}
	err = yaml.Unmarshal(d, &c)
	return
}

// 保存复读姬配置
func (c config) save(ctx *zero.Ctx) error {
	data, err := yaml.Marshal(c)
	if nil != err {
		kitten.SendWithImageFail(ctx, fmt.Errorf(`保存复读姬配置文件错误喵！%w`, err))
	}
	return getPath(configFile).Write(data)
}

// 获取路径
func getPath(name string) core.Path {
	return core.FilePath(engine.DataFolder(), name)
}
