package track

import (
	"fmt"
	"strconv"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/PuerkitoBio/goquery"
)

const cwm platform = `刺猬猫阅读`

// 小说网页信息获取
func (nv *novel) initCWM(bookID string) error {
	// 初始化小说平台
	nv.platform = cwm.String()
	// 向小说传入书号
	nv.id = bookID
	// 生成链接
	nv.url = `https://www.ciweimao.com/book/` + nv.id
	// 获取小说网页，失败则返回
	doc, err := fetchHTML(nv.url)
	if nil != err {
		return err
	}
	if `刺猬猫` == doc.Find(`title`).Text() {
		return errStatus(nv.url, bookUnreachable)
	}
	// 获取小说信息
	bookInfo := doc.Find(`div.book-info`)
	// 获取作者
	var (
		w = bookInfo.Find(`span`)
		n = w.Find(`a`).Eq(0)
	)
	nv.writer = n.Text()
	w.Eq(0).Remove()
	// 获取书名
	nv.name = bookInfo.Find(`h1.title`).Text()
	// 获取标签
	tagList := bookInfo.Find(`span.label`)
	nv.tagList = make([]string, 0, tagList.Length())
	tagList.Each(func(i int, selection *goquery.Selection) {
		nv.tagList = append(nv.tagList, core.CleanAll(selection.Text(), false))
	})
	// 获取小说状态
	nv.status = bookInfo.Find(`p.update-state`).Text()
	// 获取小说成绩
	bookGrade := bookInfo.Find(`p.book-grade`).Find(`b`)
	// 获取小说点击
	nv.hitNum = bookGrade.Eq(0).Text()
	// 获取小说收藏
	nv.collection = bookGrade.Eq(1).Text()
	// 获取小说字数
	nv.wordNum = bookGrade.Eq(2).Text()
	// 获取上架状态
	nv.isVIP = `订阅` == bookInfo.Find(`a.btn`).Eq(2).Text()
	// 获取项目
	i := doc.Find(`div.book-desc`).Find(`p`)
	nv.item = i.Text()
	i.Remove()
	// 获取简述
	nv.introduce = core.CleanAll(doc.Find(`div.book-desc`).Text(), true)
	// 获取小说数据
	property := doc.Find(`div.book-property`).Find(`span`)
	// 获取小说类别
	nv.theme = property.Eq(4).Find(`i`).Text()
	// 头像链接是否存在
	var he bool
	// 获取头像链接，失败时使用报错图片
	if nv.headURL, he = doc.Find(`div.author-info`).Find(`img.lazyload`).Attr(`data-original`); !he {
		nv.headURL = imgErr
		kitten.Warn(`头像链接获取失败喵！`)
	}
	// 封面链接是否存在
	var ce bool
	// 获取封面，失败时使用报错图片
	if nv.coverURL, ce = doc.Find(`a.cover`).Find(`img.lazyload`).Attr(`data-original`); !ce {
		nv.coverURL = imgErr
		kitten.Warn("封面链接获取失败喵！")
	}
	// 获取新章节链接
	nC, eC := doc.Find(`h3.tit`).Find(`a`).Eq(0).Attr(`href`)
	if !eC {
		// 如果新章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return errStatus(nv.url, noChapterURL)
	}
	// 从章节池初始化章节，向章节传入本书链接
	nv.newChapter = *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&nv.newChapter)
	nv.newChapter.bookURL = nv.url
	// 加载新章节
	if err := nv.newChapter.initCWM(nC); nil != err {
		return err
	}
	return nil
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
	doc, err := fetchHTML(cp.url)
	if nil != err {
		return err
	}
	if `刺猬猫` == doc.Find(`title`).Text() {
		return errStatus(url, chapterUnreachable)
	}
	hd := doc.Find(`div.read-hd`)
	// 获取章节标题
	cp.title = hd.Find(`h1.chapter`).Text()
	ps := hd.Find(`p`).Find(`span`)
	// 获取更新时间
	if 15 < len(ps.Eq(2).Text()) {
		if cp.update, err = parseTime(ps.Eq(2).Text()[15:], cwm); nil != err {
			kitten.Warnln(`时间转换出错喵！`, err)
		}
	}
	// 获取新章节字数
	if 9 < len(ps.Eq(-1).Text()) {
		if cp.wordNum, err = strconv.Atoi(ps.Eq(-1).Text()[9:]); nil != err {
			kitten.Warnln(`转换新章节字数失败了喵！`, err)
		}
	}
	var (
		rp = doc.Find(`div.book-read-page`).Find(`a`)
		ok bool
	)
	// 获取上一章链接
	if cp.lastURL, ok = rp.Eq(0).Attr(`href`); !ok {
		kitten.Debugln(url, `上一章链接获取失败喵！`) // 刺猬猫阅读的上一章链接在首章不具备跳转功能，因此获取失败不应该警告
	}
	// 获取下一章链接
	if cp.nextURL, ok = rp.Eq(2).Attr(`href`); !ok {
		kitten.Debugln(url, `下一章链接获取失败喵！`) // 刺猬猫阅读的下一章链接在末章不具备跳转功能，因此获取失败不应该警告
	}
	// 获取付费状态
	if id, ok := doc.Find(`div.read-bd`).Attr(`id`); ok {
		switch id {
		case `J_BookRead`:
			cp.isVIP = false
		case `J_ImgRead`:
			cp.isVIP = true
		default:
			return errStatus(cp.url, vipChapterException)
		}
	}
	return nil
}

// 用关键词搜索书号
func (key keyWord) findCWMBookID() (string, error) {
	doc, err := fetchHTML(fmt.Sprint(`https://www.ciweimao.com/get-search-book-list/0-0-0-0-0-0/全部/`, key, `/1`))
	if nil != err {
		return ``, err
	}
	href, ok := doc.Find(`a.cover`).Eq(0).Attr(`href`)
	if !ok {
		return ``, notFound(key)
	}
	return href[30:], nil
}
