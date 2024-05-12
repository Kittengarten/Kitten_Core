package track

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type stringCount int // 用于判断的字数

const (
	sf             platform    = `SF轻小说`
	sfHost                     = `https://book.sfacg.com`
	sfURL                      = sfHost + `/Novel/`
	sfDateTime                 = `2006/1/2 15:04:05`
	bookStrings    stringCount = 43
	chapterStrings stringCount = 38
)

// 小说网页信息获取
func (nv *novel) initSF(bookID string) error {
	// 初始化小说平台
	nv.platform = sf.String()
	// 向小说传入书号
	nv.id = bookID
	// 生成链接
	nv.url = sfURL + nv.id + `/`
	// 获取小说网页，失败则返回
	doc, err := fetchNode(nv.url)
	if nil != err {
		return err
	}
	if err = mayExist(doc, nv.url, bookStrings); nil != err {
		return err
	}
	// 获取书名
	nv.name = core.InnerText(doc, `//h1[@class="title"]/span[@class="text"]`)
	// 获取小说版权状态与项目
	nv.getNovelRightItem(doc)
	// 获取作者
	nv.writer = core.InnerText(doc, `//div[@class="author-name"]/span`)
	// 获取头像链接
	nv.headURL = core.InnerText(doc, `//div[@class="author-mask"]//img/@src`)
	// 小说详细信息
	textRow := htmlquery.Find(doc, `//div[@class="text-row"]/span`)
	// 获取类型
	nv.theme = strings.TrimPrefix(htmlquery.InnerText(textRow[0]), `类型：`)
	// 获取小说字数信息
	nv.wordNum = core.MidText(`字数：`, `字[`, htmlquery.InnerText(textRow[1]))
	// 获取状态
	nv.status = core.MidText(`[`, `]`, htmlquery.InnerText(textRow[1]))
	// 获取点击
	nv.hitNum = strings.TrimPrefix(htmlquery.InnerText(textRow[2]), `点击：`)
	// 获取简述
	nv.introduce = core.InnerText(doc, `//p[@class="introduce"]`)
	// 获取移动版简述
	if introduceMobile, err := nv.getIntroduce(); nil == err &&
		len(introduceMobile) >= len(nv.introduce) {
		nv.introduce = introduceMobile
	}
	// 获取收藏
	nv.collection = strings.TrimPrefix(
		core.InnerText(doc, `//div[@id="BasicOperation"]/a[3]`), `收藏 `)
	// 获取标签
	for _, t := range htmlquery.Find(doc,
		`//li[starts-with(@class,"tag")]/a/span[@class="text"]`) {
		nv.tagList = append(nv.tagList, core.CleanAll(htmlquery.InnerText(t), false))
	}
	// 获取封面链接
	nv.coverURL = core.InnerText(doc, `//div[@class="figure"]//img/@src`)
	// 获取预览
	nv.preview = strings.TrimPrefix(core.CleanAll(strings.ReplaceAll(core.InnerText(
		doc, `//div[@class="chapter-info"]/p`), `　　`, "\n"), true), "\n")
	// 获取新章节链接
	newChapter, err := htmlquery.Query(doc, `//div[@class="chapter-info"]/h3/a/@href`)
	if nil != err {
		// 如果新章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return fmt.Errorf(`新章节链接错误：%w%w`, errStatus(nv.url, noChapterURL), err)
	}
	// 从章节池初始化章节，向章节传入本书链接
	nv.newChapter = *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&nv.newChapter)
	nv.newChapter.bookURL = nv.url
	// 加载新章节
	if err := nv.newChapter.initSF(
		sfHost + htmlquery.InnerText(newChapter)); nil != err {
		return err
	}
	// 如果是上架
	if `VIP` != nv.right {
		return nil
	}
	// 尝试获取新公众章节链接
	newChapterFreeNode, err := htmlquery.Query(doc, `//div[@class="chapter-info"]/div/a/@href`)
	if nil != err {
		// 如果新公众章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return fmt.Errorf(`新公众章节链接错误：%w%w`, errStatus(nv.url, noChapterURL), err)
	}
	// 从章节池初始化章节，向章节传入本书链接
	newChapterFree := *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&newChapterFree)
	newChapterFree.bookURL = nv.url
	// 加载最新公众章节
	if err := newChapterFree.initSF(
		sfHost + htmlquery.InnerText(newChapterFreeNode)); nil != err {
		return err
	}
	// 如果最新公众章节比最新章节新，则以最新公众章节为准
	if newChapterFree.update.After(nv.newChapter.update) {
		nv.newChapter = newChapterFree
	}
	return nil
}

// 章节信息获取
func (cp *chapter) initSF(url string) error {
	// 防止章节炸了导致获取章节跳转引发 panic
	if url+`/` == cp.bookURL {
		return errStatus(url, chapterURLException)
	}
	// 向章节传入链接
	cp.url = url
	// 获取章节网页，失败则返回
	doc, err := fetchNode(cp.url)
	if nil != err {
		return err
	}
	if err = mayExist(doc, cp.url, bookStrings); nil != err {
		return err
	}
	// 获取章节标题
	cp.title = core.InnerText(doc, `//h1[@class="article-title"]`)
	// 获取更新时间
	cp.update, err = parseTime(strings.TrimPrefix(
		core.InnerText(doc, `//div[@class="article-desc"]/span[2]`), `更新时间：`), sf)
	if nil != err {
		return err
	}
	// 获取新章节字数
	cp.wordNum, err = strconv.Atoi(strings.TrimPrefix(
		core.InnerText(doc, `//div[@class="article-desc"]/span[3]`), `字数：`))
	if nil != err {
		return err
	}
	// 获取上一章链接
	cp.lastURL = sfHost +
		core.InnerText(doc, `//div[@id="article"]/div[@class="fn-btn"]/a[1]/@href`)
	// 获取下一章链接
	cp.nextURL = sfHost +
		core.InnerText(doc,
			`//div[@id="article"]/div[@class="fn-btn"]/a[2]/@href`)
	// 获取付费状态
	cp.isVIP = strings.Contains(url, `vip`)
	return nil
}

// 用关键词搜索书号
func (key keyword) findSFBookID() (string, error) {
	doc, err := fetchNode(fmt.Sprint(`http://s.sfacg.com/?Key=`, key, `&S=1&SS=0`))
	if nil != err {
		return ``, err
	}
	href := htmlquery.FindOne(doc,
		`//a[@id="SearchResultList1___ResultList_LinkInfo_0"]/@href`)
	if nil == href {
		return ``, notFound(key)
	}
	return strings.TrimPrefix(htmlquery.InnerText(href), sfURL), nil
}

// 获取移动版简述
func (nv *novel) getIntroduce() (string, error) {
	doc, err := fetchNode(`https://m.sfacg.com/b/` + nv.id + `/`)
	if nil != err {
		return ``, err
	}
	if err = mayExist(doc, nv.url, bookStrings); nil != err {
		return ``, err
	}
	return core.InnerText(doc, `//ul[@class="book_profile"]/li[@class="book_bk_qs1"]`), nil
}

// 判断小说或章节是否可能存在
func mayExist(doc *html.Node, url string, count stringCount) error {
	if int(count) > len(core.InnerText(doc, `//title`)) {
		return errStatus(url, bookUnreachable)
	}
	return nil
}

// 获取小说版权状态与项目
func (nv *novel) getNovelRightItem(doc *html.Node) {
	for _, tt := range htmlquery.Find(doc,
		`//h1[@class="title"]/span[starts-with(@class,"tag")]`) {
		if blue := htmlquery.FindOne(tt, `.[contains(@class,"blue")]`); `` ==
			nv.right && nil != blue {
			// 获取版权状态
			nv.right = htmlquery.InnerText(blue)
		}
		if yellow := htmlquery.FindOne(tt, `.[contains(@class,"yellow")]`); `` ==
			nv.right && nil != yellow {
			// 获取版权状态
			nv.right = htmlquery.InnerText(yellow)
		}
		if green := htmlquery.FindOne(tt, `.[contains(@class,"green")]`); `` ==
			nv.item && nil != green {
			// 获取项目
			nv.item = htmlquery.InnerText(green)
		}
	}
}
