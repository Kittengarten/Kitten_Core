package sfacg

import (
	"cmp"
	"slices"

	"github.com/Kittengarten/KittenCore/kitten"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// 加载报更
func loadConfig[T string | kitten.Path](configFile T) (c books, err error) {
	d, err := kitten.Path(configFile).Read()
	if nil != err {
		return
	}
	err = yaml.Unmarshal(d, &c)
	return
}

// 保存报更
func (c books) saveConfig(ctx *zero.Ctx, o string, n novel, e *control.Engine) {
	// 按更新时间倒序排列
	slices.SortFunc(c, func(j, i book) int { return cmp.Compare(i.UpdateTime, j.UpdateTime) })
	data, err := yaml.Marshal(c)
	if nil != err {
		zap.S().Error(err)
		kitten.SendWithImageFail(ctx, `%v`, err)
	}
	kitten.FilePath(e.DataFolder(), configFile).Write(data)
	cu <- c
	kitten.SendTextOf(ctx, false, `%s《%s》报更成功喵！`, o, n.name)
}
