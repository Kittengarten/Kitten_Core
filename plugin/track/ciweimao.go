package track

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/antchfx/htmlquery"
)

const (
	cwm     platform = `刺猬猫阅读`
	cwmHost          = `https://www.ciweimao.com`
	cwmURL           = cwmHost + `/book/`
)

// 小说网页信息获取
func (nv *novel) initCWM(bookID string) error {
	// 初始化小说平台
	nv.platform = string(cwm)
	// 向小说传入书号
	nv.id = bookID
	// 生成链接
	nv.url = cwmURL + nv.id
	// 获取小说网页，失败则返回
	doc, err := core.FetchNode(nv.url)
	if nil != err {
		return err
	}
	if `刺猬猫` == core.InnerText(doc, `//title`) {
		return errStatus(nv.url, bookUnreachable)
	}
	// 获取小说信息
	bookInfo := htmlquery.FindOne(doc, `//div[@class="book-info"]`)
	// 获取书名
	nv.name = core.InnerText(bookInfo, `/h1[@class="title"]/text()`)
	// 获取作者
	nv.writer = core.InnerText(bookInfo, `/h1[@class="title"]/span/a`)
	// 获取标签
	for _, t := range htmlquery.Find(bookInfo,
		`/p/span[starts-with(@class,"label")]/a`) {
		nv.tagList = append(nv.tagList, core.CleanAll(htmlquery.InnerText(t), false))
	}
	// 获取小说状态
	nv.status = core.InnerText(bookInfo, `/p[@class="update-state"]`)
	// 获取小说成绩
	bookGrade := htmlquery.Find(bookInfo, `/p[@class="book-grade"]/b`)
	if 3 > len(bookGrade) {
		return errStatus(nv.url, bookUnreachable)
	}
	// 获取小说点击
	nv.hitNum = htmlquery.InnerText(bookGrade[0])
	// 获取小说收藏
	nv.collection = htmlquery.InnerText(bookGrade[1])
	// 获取小说字数
	nv.wordNum = htmlquery.InnerText(bookGrade[2])
	// 获取项目
	if item := htmlquery.FindOne(doc, `//div[starts-with(@class,"book-desc")]/p`); nil != item {
		nv.item = htmlquery.InnerText(item)
	}
	// 获取简述
	var s strings.Builder
	for _, i := range htmlquery.Find(doc, `//div[starts-with(@class,"book-desc")]/text()`) {
		s.WriteString(htmlquery.InnerText(i))
	}
	nv.introduce = core.CleanAll(s.String(), true)
	// 获取小说数据
	property := htmlquery.Find(doc, `//div[starts-with(@class,"book-property")]/span/i`)
	if 9 > len(property) {
		return errStatus(nv.url, bookStatusException)
	}
	// 获取上架状态
	nv.right = htmlquery.InnerText(property[0])
	// 获取小说类别
	nv.theme = htmlquery.InnerText(property[4])
	// 获取头像链接
	nv.headURL = core.InnerText(doc, `//div[@class="author-info"]//img/@data-original`)
	// 获取封面
	nv.coverURL = core.InnerText(doc, `//a[@class="cover"]//img/@data-original`)
	// 获取新章节链接
	newChapter := htmlquery.FindOne(doc, `//h3[@class="tit"]/a[@target]/@href[1]`)
	if nil == newChapter {
		// 如果新章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return fmt.Errorf(`新章节链接错误：%w`, errStatus(nv.url, noChapterURL))
	}
	// 从章节池初始化章节，向章节传入本书链接
	nv.newChapter = *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&nv.newChapter)
	nv.newChapter.bookURL = nv.url
	// 加载新章节
	return nv.newChapter.initCWM(htmlquery.InnerText(newChapter))
}

// 章节信息获取
func (cp *chapter) initCWM(url string) error {
	// 防止章节炸了导致获取章节跳转引发 panic
	if url == cp.bookURL {
		return errStatus(url, chapterURLException)
	}
	// 向章节传入链接
	cp.url = url
	// 获取章节网页，失败则返回
	doc, err := core.FetchNode(cp.url)
	if nil != err {
		return err
	}
	if `刺猬猫` == core.InnerText(doc, `//title`) {
		return errStatus(cp.url, chapterUnreachable)
	}
	// 获取章节标题
	cp.title = core.InnerText(doc, `//div[@class="read-hd"]/h1[@class="chapter"]`)
	// 获取更新时间
	cp.update, err = parseTime(strings.TrimPrefix(
		core.InnerText(doc, `//div[@class="read-hd"]/p/span[3]`), `更新时间：`), cwm)
	if nil != err {
		return err
	}
	// 获取新章节字数
	cp.wordNum, err = strconv.Atoi(strings.TrimPrefix(
		core.InnerText(doc, `//div[@class="read-hd"]/p/span[5]`), `字数：`))
	if nil != err {
		return err
	}
	// 获取上一章链接
	last := htmlquery.FindOne(doc,
		`//div[@class="book-read-page"]/a[@id="J_BtnPagePrev"]/@href`)
	if nil != last {
		cp.lastURL = htmlquery.InnerText(last)
	}
	// 获取下一章链接
	next := htmlquery.FindOne(doc,
		`//div[@class="book-read-page"]/a[@id="J_BtnPageNext"]/@href`)
	if nil != next {
		cp.nextURL = htmlquery.InnerText(next)
	}
	// 获取付费状态
	switch core.InnerText(doc, `//div[@class="read-bd"]/@id`) {
	case `J_BookRead`:
		cp.isVIP = false
	case `J_ImgRead`:
		cp.isVIP = true
	default:
		return errStatus(cp.url, vipChapterException)
	}
	return nil
}

// 用关键词搜索书号
func (key keyword) findCWMBookID() (string, error) {
	doc, err := core.FetchNode(fmt.Sprint(
		cwmHost+`/get-search-book-list/0-0-0-0-0-0/全部/`, key, `/1`))
	if nil != err {
		return ``, err
	}
	href := htmlquery.FindOne(doc, `//a[@class="cover"]/@href`)
	if nil == href {
		return ``, notFound(key)
	}
	return strings.TrimPrefix(htmlquery.InnerText(href), cwmURL), nil
}
