package track

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/PuerkitoBio/goquery"
)

const (
	sf         platform = `SF轻小说`
	sfDateTime          = `2006/1/2 15:04:05`
)

// 小说网页信息获取
func (nv *novel) initSF(bookID string) error {
	// 初始化小说平台
	nv.platform = sf.String()
	// 向小说传入书号
	nv.id = bookID
	// 生成链接
	nv.url = `https://book.sfacg.com/Novel/` + nv.id + `/`
	// 获取小说网页，失败则返回
	doc, err := fetchHTML(nv.url)
	if nil != err {
		return err
	}
	if err = nv.bookMayExist(doc); nil != err {
		return err
	}
	// 获取书名
	nv.name = doc.Find(`h1.title`).Find(`span.text`).Text()
	// 获取上架状态
	nv.isVIP = `VIP` == doc.Find(`h1.title`).Find(`span.tag.blue`).Text()
	// 获取项目
	nv.item = doc.Find(`h1.title`).Find(`span.tag.green`).Text()
	// 获取作者
	nv.writer = doc.Find(`div.author-name`).Find(`span`).Text()
	// 头像链接是否存在
	var he bool
	// 获取头像链接，失败时使用报错图片
	if nv.headURL, he = doc.Find(`div.author-mask`).Find(`img`).Attr(`src`); !he {
		nv.headURL = imgErr
		kitten.Warn(`头像链接获取失败喵！`)
	}
	// 小说详细信息
	textRow := doc.Find(`div.text-row`).Find(`span`)
	// 获取类型
	if 9 < len(textRow.Eq(0).Text()) {
		nv.theme = textRow.Eq(0).Text()[9:]
	} else {
		kitten.Warn(`获取类型错误喵！`)
	}
	// 获取点击
	if 9 < len(textRow.Eq(2).Text()) {
		nv.hitNum = textRow.Eq(2).Text()[9:]
	} else {
		kitten.Warn(`获取点击错误喵！`)
	}
	// 获取更新时间
	if 9 < len(textRow.Eq(3).Text()) {
		nv.newChapter.update, err = parseTime(textRow.Eq(3).Text()[9:], sf)
		if nil != err {
			kitten.Warnln(`时间转换出错喵！`, err)
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
	// 获取手机版简述
	if introduceMobile, err := nv.getIntroduce(); nil == err &&
		len(introduceMobile) >= len(nv.introduce) {
		nv.introduce = introduceMobile
	}
	// 获取标签
	tagList := doc.Find(`ul.tag-list`).Find(`a`).Find(`span.text`)
	nv.tagList = make([]string, 0, tagList.Length())
	tagList.Each(func(i int, selection *goquery.Selection) {
		nv.tagList = append(nv.tagList, selection.Text())
	})
	// 封面链接是否存在
	var ce bool
	// 获取封面，失败时使用报错图片
	if nv.coverURL, ce = doc.Find(`div.figure`).Find(`img`).Eq(0).Attr(`src`); !ce {
		nv.coverURL = imgErr
		kitten.Warn("封面链接获取失败喵！")
	}
	// 获取收藏
	if 7 < len(doc.Find(`#BasicOperation`).Find(`a`).Eq(2).Text()) {
		nv.collection = doc.Find(`#BasicOperation`).Find(`a`).Eq(2).Text()[7:]
	}
	// 获取预览
	nv.preview = core.CleanAll(doc.Find(`div.chapter-info`).Find(`p`).Text(), true)
	// 获取新章节链接
	nC, eC := doc.Find(`div.chapter-info`).Find(`h3`).Find(`a`).Attr(`href`)
	if !eC {
		// 如果新章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return fmt.Errorf(`新章节链接错误：%w`, errStatus(nv.url, noChapterURL))
	}
	// 从章节池初始化章节，向章节传入本书链接
	nv.newChapter = *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&nv.newChapter)
	nv.newChapter.bookURL = nv.url
	// 加载新章节
	if err := nv.newChapter.initSF(`https://book.sfacg.com` + nC); nil != err {
		return err
	}
	// 如果是上架
	if nv.isVIP {
		// 尝试获取新公众章节链接
		nCF, eCF := doc.Find(`div.chapter-info`).Find(`div`).Find(`a`).Attr(`href`)
		if !eCF {
			// 如果新公众章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
			return fmt.Errorf(`新公众章节链接错误：%w`, errStatus(nv.url, noChapterURL))
		}
		// 从章节池初始化章节，向章节传入本书链接
		newChapterFree := *chapterPool.Get().(*chapter)
		defer chapterPool.Put(&newChapterFree)
		newChapterFree.bookURL = nv.url
		// 加载最新公众章节
		if err := newChapterFree.initSF(`https://book.sfacg.com` + nCF); nil != err {
			return err
		}
		// 如果最新公众章节比最新章节新，则以最新公众章节为准
		if newChapterFree.update.After(nv.newChapter.update) {
			nv.newChapter = newChapterFree
		}
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
	doc, err := fetchHTML(cp.url)
	if nil != err {
		return err
	}
	if title := doc.Find(`title`).Text(); 38 > len(title) {
		return errStatus(url, chapterUnreachable)
	}
	desc := doc.Find(`div.article-desc`).Find(`span`)
	// 获取更新时间
	if 15 < len(desc.Eq(1).Text()) {
		if cp.update, err = parseTime(desc.Eq(1).Text()[15:], sf); nil != err {
			kitten.Warn(`时间转换出错喵！`, err)
		}
	}
	// 获取新章节字数
	if 9 < len(desc.Eq(2).Text()) {
		if cp.wordNum, err = strconv.Atoi(desc.Eq(2).Text()[9:]); nil != err {
			kitten.Warnln(`转换新章节字数失败了喵！`, err)
		}
	}
	// 获取章节标题
	cp.title = doc.Find(`h1.article-title`).Text()
	var ok bool
	// 获取上一章链接
	if cp.lastURL, ok = doc.Find(`div.article`).Find(`div.fn-btn`).Eq(-1).Find(`a`).Eq(0).Attr(`href`); !ok {
		kitten.Warnln(url, ` 上一章链接获取失败喵！`) // SF 轻小说的上一章链接在首章会跳转书首页，因此获取失败应该警告
	}
	cp.lastURL = `https://book.sfacg.com` + cp.lastURL
	// 获取下一章链接
	if cp.nextURL, ok = doc.Find(`div.article`).Find(`div.fn-btn`).Eq(-1).Find(`a`).Eq(1).Attr(`href`); !ok {
		kitten.Warnln(url, ` 下一章链接获取失败喵！`) // SF 轻小说的下一章链接在末章会跳转书首页，因此获取失败应该警告
	}
	cp.nextURL = `https://book.sfacg.com` + cp.nextURL
	// 获取付费状态
	cp.isVIP = strings.Contains(url, `vip`)
	return nil
}

// 用关键词搜索书号
func (key keyWord) findSFBookID() (string, error) {
	doc, err := fetchHTML(fmt.Sprint(`http://s.sfacg.com/?Key=`, key, `&S=1&SS=0`))
	if nil != err {
		return ``, err
	}
	href, ok := doc.Find(`#SearchResultList1___ResultList_LinkInfo_0`).Attr(`href`)
	if !ok {
		return ``, notFound(key)
	}
	return href[29:], nil
}

// 获取移动版简述
func (nv *novel) getIntroduce() (string, error) {
	doc, err := fetchHTML(`https://m.sfacg.com/b/` + nv.id)
	if nil != err {
		return ``, err
	}
	return doc.Find(`ul.book_profile`).Find(`li.book_bk_qs1`).Text(), nv.bookMayExist(doc)
}

// 判断小说是否可能存在
func (nv *novel) bookMayExist(doc *goquery.Document) error {
	if 43 <= len(doc.Find(`title`).Text()) {
		return nil
	}
	return errStatus(nv.url, bookUnreachable)
}
