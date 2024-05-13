package track

import (
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	"gopkg.in/yaml.v3"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// 加载报更
func loadConfig(configFile string) (c books, err error) {
	d, err := getPath(configFile).Read()
	if nil != err {
		return
	}
	err = yaml.Unmarshal(d, &c)
	return
}

// 保存报更
func (c *books) saveConfig(ctx *zero.Ctx) (err error) {
	c.SortByUpdate()
	data, err := yaml.Marshal(c)
	if nil != err {
		kitten.SendWithImageFail(ctx, errSave, err)
	}
	if err = getPath(configFile).Write(data); nil != err {
		return
	}
	cu <- *c
	return
}

// 获取路径
func getPath(name string) core.Path {
	return core.FilePath(engine.DataFolder(), name)
}

// 按更新时间倒序排列小说
func (c *books) SortByUpdate() {
	slices.SortFunc(*c, func(j, i book) int {
		ti, err := parseTime(i.UpdateTime, ``)
		if nil != err {
			// 时间转换出错则不处理
			if `` != i.UpdateTime {
				kitten.Error(err)
			}
			return 0
		}
		tj, err := parseTime(j.UpdateTime, ``)
		if nil != err {
			// 时间转换出错则不处理
			if `` != j.UpdateTime {
				kitten.Error(err)
			}
			return 0
		}
		return ti.Compare(tj)
	})
}

// 时间解析，匹配不到支持的平台时使用默认时间格式
func parseTime(str string, p platform) (time.Time, error) {
	if `` == str {
		return time.Time{}, nil
	}
	switch p {
	case sf:
		return time.Parse(sfDateTime, str)
	case cwm:
		return time.Parse(time.DateTime, str)
	default:
		return time.Parse(core.Layout, str)
	}
}

// String 实现 fmt.Stringer
func (key keyword) String() string {
	return string(key)
}

// String 实现 fmt.Stringer
func (p platform) String() string {
	return string(p)
}
