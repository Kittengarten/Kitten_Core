package sfacg

import "time"

type (
	// 一本小说的数据集
	novel struct {
		id         string  // 小说书号
		url        string  // 小说网址
		name       string  // 小说书名
		isVIP      bool    // 是否上架
		writer     string  // 作者昵称
		hitNum     string  // 小说点击
		wordNum    string  // 小说字数
		preview    string  // 章节预览
		headURL    string  // 头像网址
		coverURL   string  // 封面网址
		collection string  // 小说收藏
		newChapter chapter // 章节信息
		theme      string  // 小说类型（主题）
		introduce  string  // 小说简述
		status     string  // 小说状态
		// tagList    []string // 标签列表
	}

	// 一个章节的数据集
	chapter struct {
		bookURL string    // 本书网址
		url     string    // 章节网址
		isVIP   bool      // 是否付费章节
		update  time.Time // 更新时间
		title   string    // 章节名称
		wordNum int       // 章节字数
		lastURL string    // 上章网址
		nextURL string    // 下章网址
	}

	// 章节之间比较的数据集
	compare struct {
		times   int           // 当日更新次数
		wordNum int           // 当日更新字数
		timeGap time.Duration // 距离上次更新的时间差
	}

	// 多项小说报更项目的数据集组成的切片
	books []book

	// 小说报更项目的数据集
	book struct {
		BookID     string  `yaml:"bookid"`     // 报更书号
		BookName   string  `yaml:"bookname"`   // 报更书名
		GroupID    []int64 `yaml:"groupid"`    // 书友群号，负数代表私聊 QQ
		RecordURL  string  `yaml:"recordurl"`  // 上次更新链接
		UpdateTime string  `yaml:"updatetime"` // 上次更新时间
	}

	keyWord string // 搜索关键词
)
