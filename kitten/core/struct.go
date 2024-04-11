package core

const (
	// Empty YAML 空数组
	Empty = `[]`
	// Layout 日期时间格式
	Layout = `2006.1.2	♥	15:04:05`
	// MaxInt 最大 int
	MaxInt = int(^uint(0) >> 1)
	// 平台位数
	PlatformBits = 32 << (^uint(0) >> 63)
)

type (
	// Path 是一个表示文件路径的字符串
	Path string

	// Choicers 是由随机项目的抽象接口组成的切片
	Choicers []interface {
		GetID() int             // 该项目的 ID
		GetInformation() string // 该项目的信息
		GetChance() int         // 该项目的权重
	}

	// HTTPErr 是一个表示 HTTP 错误的结构体
	HTTPErr struct {
		URL        string // 请求的 URL
		Method     string // 请求的方法
		StatusCode int    // HTTP 状态码
	}
)
