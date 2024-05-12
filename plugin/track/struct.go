package track

import "time"

type (
	// 一本小说
	novel struct {
		platform   string  // 小说平台
		id         string  // 小说书号
		name       string  // 小说书名
		url        string  // 小说网址
		writerInfo         // 作者信息
		data               // 小说数据
		info               // 小说信息
		newChapter chapter // 章节信息
		compare            // 章节之间比较
	}

	// 作者信息
	writerInfo struct {
		writer  string // 作者昵称
		headURL string // 头像网址
	}

	// 小说数据
	data struct {
		right      string // 版权状态
		collection string // 小说收藏
		hitNum     string // 小说点击
		wordNum    string // 小说字数
	}

	// 小说信息
	info struct {
		coverURL  string   // 封面网址
		preview   string   // 章节预览
		theme     string   // 小说类型（主题）
		introduce string   // 小说简述
		status    string   // 小说状态
		item      string   // 小说参加的项目
		tagList   []string // 标签列表
	}

	// 一个章节
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
		times        int           // 当日更新次数
		todayWordNum int           // 当日更新字数
		timeGap      time.Duration // 距离上次更新的时间差
	}

	// 多项小说报更项目的数据集组成的切片
	books []book

	// 小说报更项目的数据集
	book struct {
		Platform   string  `yaml:"platform"`   // 报更平台
		BookID     string  `yaml:"bookid"`     // 报更书号
		BookName   string  `yaml:"bookname"`   // 报更书名
		GroupID    []int64 `yaml:"groupid"`    // 书友群号，负数代表私聊 QQ
		RecordURL  string  `yaml:"recordurl"`  // 上次更新链接
		UpdateTime string  `yaml:"updatetime"` // 上次更新时间
	}

	platform string // 平台

	keyword string // 搜索关键词
)
