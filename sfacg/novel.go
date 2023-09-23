package sfacg

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

var (
	// 图片路径
	imagePath = kitten.Path(kitten.GetMainConfig().Path + `image/`)
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
func (nv *novel) init(bookID string) error {
	// 向小说传入书号
	nv.id = bookID
	// 生成链接
	nv.url = fmt.Sprint(`https://book.sfacg.com/Novel/`, nv.id)
	// 从章节池初始化章节，向章节传入本书链接
	nv.newChapter = *chapterPool.Get().(*chapter)
	nv.newChapter.bookURL = nv.url
	defer chapterPool.Put(&nv.newChapter)
	// 获取 HTTP 响应，失败则返回
	res, err := http.Get(nv.url)
	if nil != err {
		return err
	}
	defer res.Body.Close()
	// 小说网页
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if title := doc.Find(`title`).Text(); 43 > len(title) {
		// 书无了则返回
		if nil != err {
			return err
		}
		return fmt.Errorf(`书号 %s 没有喵！`, bookID)
	}
	// 获取书名
	nv.name = doc.Find(`h1.title`).Find(`span.text`).Text()
	// 获取作者
	nv.writer = doc.Find(`div.author-name`).Find(`span`).Text()
	// 头像链接是否存在
	var he bool
	// 获取头像链接，失败时使用报错图片
	if nv.headURL, he = doc.Find(`div.author-mask`).Find(`img`).Attr(`src`); !he {
		nv.headURL = fmt.Sprint(`file://`, imagePath, `no.png`)
		zap.S().Warn(`头像链接获取失败喵！`)
	}
	// 小说详细信息
	textRow := doc.Find(`div.text-row`).Find(`span`)
	// 获取类型
	if 9 < len(textRow.Eq(0).Text()) {
		nv.theme = textRow.Eq(0).Text()[9:]
	} else {
		zap.S().Warn(`获取类型错误喵！`)
	}
	// 获取点击
	if 9 < len(textRow.Eq(2).Text()) {
		nv.hitNum = textRow.Eq(2).Text()[9:]
	} else {
		zap.S().Warn(`获取点击错误喵！`)
	}
	// 获取更新时间
	if 9 < len(textRow.Eq(3).Text()) {
		nv.newChapter.update, err = time.Parse(`2006/1/2 15:04:05`, textRow.Eq(3).Text()[9:])
		if nil != err {
			zap.S().Warn(`时间转换出错喵！`, err)
		}
	}
	// 获取小说字数信息
	var (
		wordNumInfo = textRow.Eq(1).Text()
		lw          = len(wordNumInfo)
	)
	// 获取字数
	if 9 < lw {
		nv.wordNum = wordNumInfo[9 : lw-14]
	}
	// 获取状态
	if 11 < lw {
		nv.status = wordNumInfo[lw-11:]
	}
	// 获取简述
	nv.introduce = doc.Find(`p.introduce`).Text()
	// // 获取标签
	// doc.Find(`ul.tag-list`).Find(`a`).Find(`span.text`).Each(func(i int, selection *goquery.Selection) {
	// 	nv.tagList[i] = selection.Text()
	// })
	// 封面链接是否存在
	var ce bool
	// 获取封面，失败时使用报错图片
	if nv.coverURL, ce = doc.Find(`div.figure`).Find(`img`).Eq(0).Attr(`src`); !ce {
		nv.coverURL = fmt.Sprint(`file://`, imagePath, `no.png`)
		zap.S().Warn("封面链接获取失败喵！")
	}
	// 获取收藏
	if 7 < len(doc.Find(`#BasicOperation`).Find(`a`).Eq(2).Text()) {
		nv.collection = doc.Find(`#BasicOperation`).Find(`a`).Eq(2).Text()[7:]
	}
	// 获取预览
	nv.preview = kitten.CleanAll(doc.Find(`div.chapter-info`).Find(`p`).Text(), true)
	// 获取新章节链接
	nC, eC := doc.Find(`div.chapter-info`).Find(`h3`).Find(`a`).Attr(`href`)
	if !eC {
		// 如果新章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return errors.New(nv.url + `获取更新链接失败了喵！`)
	}
	// 获取上架状态
	nv.isVIP = strings.Contains(nC, `vip`)
	// 加载新章节
	return nv.newChapter.init(fmt.Sprint(`https://book.sfacg.com`, nC))
}

// 章节信息获取
func (cp *chapter) init(url string) error {
	// 防止章节炸了导致获取章节跳转引发 panic
	if url == cp.bookURL {
		return fmt.Errorf(`%s 更新异常喵！`, url)
	}
	// 向章节传入链接
	cp.url = url
	// 获取 HTTP 响应，失败则返回
	res, err := http.Get(cp.url)
	if nil != err {
		return err
	}
	defer res.Body.Close()
	// 章节网页
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if title := doc.Find(`title`).Text(); 38 > len(title) {
		// 章节无了则返回
		if nil != err {
			return err
		}
		return fmt.Errorf(`章节 %s 没有喵！`, cp.url)
	}
	desc := doc.Find(`div.article-desc`).Find(`span`)
	// 获取更新时间
	if 15 < len(desc.Eq(1).Text()) {
		cp.update, err = time.Parse(`2006/1/2 15:04:05`, desc.Eq(1).Text()[15:])
		if nil != err {
			zap.S().Warn(`时间转换出错喵！`, err)
		}
	}
	// 获取新章节字数
	if 9 < len(desc.Eq(2).Text()) {
		cp.wordNum, err = strconv.Atoi(desc.Eq(2).Text()[9:])
	}
	if nil != err {
		zap.S().Warn(`转换新章节字数失败了喵！`, err)
	}
	// 获取章节标题
	cp.title = doc.Find(`h1.article-title`).Text()
	var ok bool
	// 获取上一章链接
	if cp.lastURL, ok = doc.Find(`div.fn-btn`).Eq(-1).Find(`a`).Eq(0).Attr(`href`); !ok {
		zap.S().Warn(url, `上一章链接获取失败喵！`)
	}
	cp.lastURL = fmt.Sprint(`https://book.sfacg.com`, cp.lastURL)
	// 获取下一章链接
	if cp.nextURL, ok = doc.Find(`div.fn-btn`).Eq(-1).Find(`a`).Eq(1).Attr(`href`); !ok {
		zap.S().Warn(url, `下一章链接获取失败喵！`)
	}
	cp.nextURL = fmt.Sprint(`https://book.sfacg.com`, cp.nextURL)
	// 获取付费状态
	cp.isVIP = strings.Contains(url, `vip`)
	return nil
}

// 与上次更新比较
func (nv *novel) makeCompare() (cm compare, err error) {
	var this, last chapter
	this = nv.newChapter
	err = last.init(this.lastURL)
	if nil != err {
		return
	}
	cm.wordNum = this.wordNum
	cm.timeGap = max(time.Second, this.update.Sub(last.update))
	for cm.times = 1; kitten.IsSameDate(last.update, this.update); cm.times++ {
		this = last
		cm.wordNum += this.wordNum
		err = last.init(this.lastURL)
		if nil != err {
			return
		}
	}
	return
}

// 小说信息
func (nv *novel) information() string {
	if `` == nv.id {
		return `获取不到书号喵！`
	}
	// var tags string // 标签
	// for k := range nv.tagList {
	// 	tags += fmt.Sprintf(`[%s]`, nv.tagList[k])
	// }
	return fmt.Sprintf(`书名：%s
书号：%s
作者：%s
【%s】%s
收藏：%s
字数：%s%s
点击：%s
更新：%s
	
%s`,
		nv.name,
		nv.id,
		nv.writer,
		nv.theme, func(v bool) string {
			if v {
				return `（上架）`
			}
			return `（免费）`
		}(nv.isVIP),
		nv.collection,
		nv.wordNum, nv.status,
		nv.hitNum,
		nv.newChapter.update.Format(kitten.Layout),
		nv.introduce)
}

// 用关键词搜索书号
func (key keyWord) findBookID() (string, error) {
	res, err := http.Get(fmt.Sprint(`http://s.sfacg.com/?Key=`, key, `&S=1&SS=0`))
	if nil != err {
		return ``, err
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if nil != err {
		return ``, err
	}
	href, ok := doc.Find(`#SearchResultList1___ResultList_LinkInfo_0`).Attr(`href`)
	if !ok {
		return ``, fmt.Errorf(`关键词【%s】找不到小说喵！`, key)
	}
	return href[29:], nil
}

// 更新信息
func (nv *novel) update() (str string, d time.Duration) {
	timeGap, todayReport := func() (string, string) {
		cm, err := nv.makeCompare()
		if nil != err {
			return `不明`, ``
		}
		d = cm.timeGap
		if d > 144*time.Hour {
			return `不明`, ``
		}
		return strings.NewReplacer(`h`, ` 小时 `, `m`, ` 分钟 `, `s`, ` 秒`).Replace(d.String()),
			fmt.Sprintf("当日第 %d 更%s", cm.times, func(t int) string {
				if 1 >= t {
					return ``
				}
				return fmt.Sprintf(`，日更 %d 字`, cm.wordNum)
			}(cm.times))
	}()
	str = fmt.Sprintf(`《%s》更新了喵～
%s
更新字数：%d 字%s
间隔时间：%s
%s`,
		nv.name,
		nv.newChapter.title,
		nv.newChapter.wordNum, func(v bool) string {
			if v {
				return `（付费）`
			}
			return `（免费）`
		}(nv.newChapter.isVIP),
		timeGap,
		todayReport)
	return
}
