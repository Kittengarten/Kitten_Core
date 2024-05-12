package track

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"
)

var (
	// 图片路径
	imagePath = core.FilePath(kitten.MainConfig().Path, `image`)
	// 小说池
	novelPool = sync.Pool{
		New: func() any {
			return new(novel)
		},
	}
	// 章节池
	chapterPool = sync.Pool{
		New: func() any {
			return new(chapter)
		},
	}
)

// 小说网页信息获取
func (nv *novel) init(p platform, bookID string) error {
	switch p {
	case sf:
		return nv.initSF(bookID)
	case cwm:
		return nv.initCWM(bookID)
	default:
		return notSupported(p)
	}
}

// 章节信息获取
func (cp *chapter) init(p platform, url string) error {
	switch p {
	case sf:
		return cp.initSF(url)
	case cwm:
		return cp.initCWM(url)
	default:
		return notSupported(p)
	}
}

// String 实现 fmt.Stringer
func (cp *chapter) String() string {
	return cp.title + `
` + cp.url
}

// String 实现 fmt.Stringer
func (nv *novel) String() string {
	if `` == nv.id {
		return `获取不到书号喵！`
	}
	var tags strings.Builder // 标签
	tags.Grow(8 * len(nv.tagList))
	for _, t := range nv.tagList {
		tags.WriteString(`[` + t + `]`)
	}
	p := platform(nv.platform)
	return `平台：` + nv.platform + `
书名：` + nv.name + `
书号：` + nv.id + `
作者：` + nv.writer + `
` + nv.url + `
【` + nv.theme + func(r string) string {
		if `` == r {
			return `】`
		}
		return `】（` + r + `）`
	}(nv.right) + func(i string) string {
		switch p {
		case sf:
			if 0 == len(i) {
				return ``
			}
			return `【` + i + `】`
		case cwm:
			return nv.item
		default:
			return ``
		}
	}(nv.item) + `
` + tags.String() + `
收藏：` + nv.collection + `
字数：` + nv.wordNum + func(i string) string {
		switch p {
		case sf, cwm:
			if 0 == len(i) {
				return ``
			}
			return `（` + i + `）`
		default:
			return ``
		}
	}(nv.status) + `
点击：` + nv.hitNum + `
更新：` + nv.newChapter.update.Format(core.Layout) + func() string {
		switch p {
		case sf:
			return `

`
		default:
			return ``
		}
	}() + nv.introduce
}

// 用关键词搜索书号
func (key keyword) findBookID(p platform) (string, error) {
	switch p {
	case sf:
		return key.findSFBookID()
	case cwm:
		return key.findCWMBookID()
	default:
		return ``, notSupported(p)
	}
}

// 与上次更新比较
func (nv *novel) makeCompare() error {
	var (
		this, last chapter
		p          = platform(nv.platform)
	)
	this = nv.newChapter
	if `` == this.lastURL {
		return errStatus(nv.url, onlyAChapter)
	}
	if err := last.init(p, this.lastURL); nil != err {
		return err
	}
	nv.todayWordNum = this.wordNum
	nv.timeGap = max(time.Second, this.update.Sub(last.update))
	for nv.times = 1; core.IsSameDate(last.update, this.update) &&
		last.lastURL != nv.url; nv.times++ {
		this = last
		nv.todayWordNum += this.wordNum
		if err := last.init(p, this.lastURL); nil != err {
			break
		}
	}
	return nil
}

// 更新信息
func (nv *novel) update() string {
	return fmt.Sprintf(`《%s》更新了喵～
%s
%s
更新字数：%d 字（%s）
间隔时间：%s`,
		nv.name,
		nv.newChapter.title,
		nv.newChapter.url,
		nv.newChapter.wordNum, func(v bool) string {
			if v {
				return `付费`
			}
			return `免费`
		}(nv.newChapter.isVIP),
		nv.todayReport(),
	)
}

// 今日报更
func (nv *novel) todayReport() string {
	if err := nv.makeCompare(); nil != err {
		return err.Error()
	}
	s, err := nv.timeGapConvert()
	if nil != err {
		return err.Error()
	}
	return s.String() + `
` + nv.todayUpdate()
}

// 距上次更新时间的时间差转换为时间间隔的结构体
func (nv *novel) timeGapConvert() (core.TimeDuration, error) {
	// 如果时间早于 2006.1.2 15:04:05
	if s, _ := parseTime(core.Layout, ``); nv.timeGap > time.Since(s) {
		return core.TimeDuration{}, errStatus(nv.url, timeException)
	}
	return core.ConvertTimeDuration(nv.timeGap), nil
}

// 今日更新信息
func (nv *novel) todayUpdate() string {
	if 1 >= nv.times {
		return fmt.Sprintf(`当日第 %d 更`, nv.times)
	}
	return fmt.Sprintf(`当日第 %d 更，日更 %d 字`, nv.times, nv.todayWordNum)
}
