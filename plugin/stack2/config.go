package stack2

import (
	"fmt"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"
	zero "github.com/wdvxdr1123/ZeroBot"

	"gopkg.in/yaml.v3"
)

var (
	// 叠猫猫配置文件名
	configFile = core.FilePath(`plugin`, replyServiceName, `config.yaml`)
	// 叠猫猫配置文件
	stackConfig, err = loadConfig(configFile)
	// 图片路径
	imagePath = core.FilePath(kitten.MainConfig().Path, replyServiceName, `image`)
	// bot 配置
	botConfig = kitten.MainConfig()
	// bot ID
	sid = botConfig.SelfID
	// 指令前缀
	p = botConfig.CommandPrefix
	// 帮助文本
	help = fmt.Sprintf(`%s%s%s %s|%s|%s|%s

摔下去会导致体重减少，压坏会导致体重增加。
没有猫猫时，有(抱枕突破所需体重/当前体重)的概率发生平地摔导致体重变为 e 倍。
清空猫堆有(抱枕突破所需体重/当前体重)的概率触发特效导致体重变为 e 倍。

抱枕、奶猫和绒布球不会导致猫猫摔下去。

猫娘能保护身边的猫猫。随着猫娘的成长，她们的能力也会越来越强。
直接在猫娘以上级别的身上叠猫猫必定不会摔下去；
猫娘萝莉以上可以保护上面一只猫猫不被压坏；
猫娘少女和成年猫娘以上能保护下面一只猫猫不摔下去，且享有分析图片特权。

压坏了别的猫猫；
被别的猫猫压坏；
叠猫猫失败摔下来；
平地摔——
这些情况需要休息 | N(0, 体重²) 小时 |（至少为 %d 小时）后，才能再次加入。`,
		p, cStack, cMeow, cIn, cView, cAnalysis, cRank,
		stackConfig.MinGapTime)
	// 吃猫猫帮助文本
	helpEat = fmt.Sprintf(`%s%s%s
需要休息 | N(0, (e*体重)²) 小时 |（至少为 %d 小时）后，才能再次加入`,
		p, cEat, cMeow,
		stackConfig.MinGapTime)
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            brief,
		Help:             help,
		PublicDataFolder: `Stack2`,
	})
)

// 加载叠猫猫配置
func loadConfig(p core.Path) (c config, err error) {
	return load[config](p, `gaptime: 1        # 每千克体重的休息时间（小时数）
mingaptime: 1     # 最小休息时间（小时数）`)
}

// 加载叠猫猫数据
func loadData(p core.Path) (d data, err error) {
	return load[data](p, core.Empty)
}

// 加载叠猫猫配置或数据，d 为默认值
func load[T config | data](p core.Path, d string) (s T, err error) {
	if err = core.InitFile(&p, d); nil != err {
		return
	}
	b, err := core.Path(p).Read()
	if nil != err {
		return
	}
	err = yaml.Unmarshal(b, &s)
	return
}

// 获取路径
func getPath(name string) core.Path {
	return core.FilePath(engine.DataFolder(), name)
}
