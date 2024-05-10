package core

const (
	Empty        = `[]`                   // Empty YAML 空数组
	Layout       = `2006.1.2	♥	15:04:05`  // Layout 日期时间格式
	MaxInt       = int(^uint(0) >> 1)     // MaxInt 最大 int
	PlatformBits = 32 << (^uint(0) >> 63) // PlatformBits 平台位数
	HoursPerDay  = 24                     // HoursPerDay 每天小时数
)
