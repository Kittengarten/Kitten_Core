package rcon

import (
	"regexp"
	"strings"
	"sync"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type (
	rcon struct {
		HOST     string // RCON 主机
		Password string // RCON 密码
	}

	item byte
)

const (
	replyServiceName = `rcon` // 插件名
	brief            = `RCON 命令执行`
	configFile       = `config.yaml` // 配置文件名
	help             = `RCON [命令]
————
私聊可用：
设置 RCON 主机 [主机]
设置 RCON 密码 [密码]`
	host     item = iota // 主机
	password             // 密码
)

var (
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             brief,
		Help:              help,
		PrivateDataFolder: replyServiceName,
	}).ApplySingle(ctxext.DefaultSingle)
	// 配置文件路径
	configPath = core.FilePath(engine.DataFolder(), configFile)
	// 读写锁
	mu sync.RWMutex
	// 设置类型
	setItem = map[item]string{
		host:     `主机`,
		password: `密码`,
	}
)

func init() {
	// RCON
	engine.OnPrefixGroup([]string{`RCON`, `rcon`}, zero.SuperUserPermission).
		SetBlock(true).Handle(command)

	// 设置 RCON 主机
	engine.OnRegex(`^设置\s*(?i)RCON\s*主机\s*(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).
		SetBlock(true).Handle(setHOST)

	// 设置 RCON 密码
	engine.OnRegex(`^设置\s*(?i)RCON\s*密码\s*(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).
		SetBlock(true).Handle(setPassword)
}

// RCON
func command(ctx *zero.Ctx) {
	mu.RLock()
	defer mu.RUnlock()
	config, err := core.Load[rcon](configPath, core.Empty)
	if nil != err {
		kitten.SendWithImageFail(ctx, `加载 RCON 配置文件错误喵！`, err)
		return
	}
	conn := &MCConn{}
	if err = conn.Open(config.HOST, config.Password); nil != err {
		kitten.SendWithImageFail(ctx, `连接 RCON 服务器错误喵！`, err)
		return
	}
	defer conn.Close()
	if err = conn.Authenticate(); nil != err {
		kitten.SendWithImageFail(ctx, `RCON 密码验证错误喵！`, err)
		return
	}
	resp, err := conn.SendCommand(kitten.GetArgs(ctx))
	if nil != err {
		kitten.SendWithImageFail(ctx, `发送 RCON 命令错误喵！`, err)
		return
	}
	if `` == resp {
		kitten.SendText(ctx, false, `命令响应为空喵！`)
		return
	}
	kitten.SendText(ctx, true, regexp.MustCompile(`§.`).
		ReplaceAllString(strings.TrimRight(resp, "\n\r"), ``))
}

// 设置 RCON 主机
func setHOST(ctx *zero.Ctx) {
	set(ctx, host)
}

// 设置 RCON 密码
func setPassword(ctx *zero.Ctx) {
	set(ctx, password)
}

func set(ctx *zero.Ctx, i item) {
	s := ctx.State["regex_matched"].([]string)[1]
	mu.Lock()
	defer mu.Unlock()
	config, err := core.Load[rcon](configPath, core.Empty)
	if nil != err {
		kitten.SendWithImageFail(ctx, `加载 RCON 配置文件错误喵！`, err)
		return
	}
	switch i {
	case host:
		config.HOST = s
	case password:
		config.Password = s
	}
	if err = core.Save[rcon](configPath, config); nil != err {
		kitten.SendWithImageFail(ctx, `保存 RCON 配置文件错误喵！`, err)
	}
	kitten.SendTextOf(ctx, false, `RCON %s设置成功喵！`, i)
}

// String 实现 fmt.Stringer
func (i item) String() string {
	return setItem[i]
}
