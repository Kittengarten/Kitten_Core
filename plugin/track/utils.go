package track

import (
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core"
)

// 保存报更
func (c *books) saveConfig() error {
	c.SortByUpdate()
	if err := core.Save(configPath, *c); nil != err {
		return err
	}
	cu <- *c
	return nil
}

// 按更新时间倒序排列小说
func (c *books) SortByUpdate() {
	slices.SortFunc(*c, func(j, i book) int {
		return i.UpdateTime.Compare(j.UpdateTime)
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
