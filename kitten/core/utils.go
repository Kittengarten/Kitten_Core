package core

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
	"unicode"

	wr "github.com/mroth/weightedrand/v2"
)

type (
	// Choicers 是由随机项目的抽象接口组成的切片
	Choicers []interface {
		GetID() int             // 该项目的 ID
		GetInformation() string // 该项目的信息
		GetChance() int         // 该项目的权重
	}

	// TimeDuration 表示时间间隔的结构体
	TimeDuration struct {
		d, h, m, s time.Duration // 天，小时，分钟，秒
	}
)

// Choose 按权重抽取一个项目的序号
func (c Choicers) Choose() (result int, err error) {
	choices := make([]wr.Choice[int, int], len(c), len(c))
	for i, ch := range c {
		item, weight := ch.GetID(), ch.GetChance()
		choices[i] = wr.Choice[int, int]{Item: item, Weight: weight}
	}
	chooser, err := wr.NewChooser(choices...)
	if nil != err {
		return -1, err
	}
	return chooser.Pick(), nil
}

// IsSameDate 判断两个时间是否在同一天
func IsSameDate(t1, t2 time.Time) bool {
	return t1.Day() == t2.Day() &&
		t1.Month() == t2.Month() &&
		t1.Year() == t2.Year()
}

/*
CleanAll 清理字符串中全部不必要内容

lf 控制是否换行
*/
func CleanAll[T ~string | ~[]rune | ~[]byte](s T, lf bool) T {
	return T(strings.Map(func(r rune) rune {
		if remove := unicode.IsControl(r) || unicode.IsSpace(r);
		// 如果不换行，移除包括换行符在内的控制字符和空白字符
		(!lf && remove) ||
			// 如果换行，移除换行符以外的控制字符和空白字符
			(lf && remove && !strings.ContainsRune("\n\r", r)) ||
			// 移除可能引发排版和显示错误的字符
			strings.ContainsRune("\u061c\u200e\u200f\u202a\u202b\u202c\u202d\u202e\u2066\u2067\u2068\u2069\ufffd", r) {
			return -1
		}
		return r
	}, string(s)))
}

/*
MidText 获取中间字符串

pre 为前缀（不包含），suf 为后缀（不包含），str 为整个字符串
*/
func MidText(pre, suf, str string) string {
	return str[func() int {
		// 截掉前缀及之前部分
		if i := strings.Index(str, pre); -1 != i {
			return i + len(pre)
		}
		return 0
	}():func() int {
		// 截掉后缀及之后部分
		if i := strings.LastIndex(str, suf); -1 != i {
			return i
		}
		return len(str)
	}()]
}

// GenerateRandomNumber 生成 count 个 [start, end) 范围的不重复的随机数
func GenerateRandomNumber(start, end, count int) ([]int, error) {
	// 范围检查
	if end <= start {
		return nil, fmt.Errorf(`上限 %d 必须大于下限 %d 喵！`, end, start)
	}
	if (end - start) < count {
		return nil, fmt.Errorf(`下限 %d 和上限 %d 之间的数字只有 %d 个，不满足 %d 个的要求喵！`, start, end, end-start, count)
	}
	if 0 >= count {
		return nil, fmt.Errorf(`个数 %d 不是正整数喵！`, count)
	}
	// 存放不重复结果的集合
	set := make(map[int]struct{}, count)
	for len(set) < count {
		// 生成随机数
		set[rand.IntN(end-start)+start] = struct{}{}
	}
	// 存放结果的切片
	nums := make([]int, 0, count)
	// 集合转换为切片
	for k := range set {
		nums = append(nums, k)
	}
	return nums, nil
}

// RandomDelay 随机阻塞等待
func RandomDelay(t time.Duration) {
	<-time.NewTimer(time.Duration(float64(t) * rand.Float64())).C
}

// ConvertTimeDuration 转换时间间隔
func ConvertTimeDuration(d time.Duration) TimeDuration {
	return TimeDuration{
		d: d / HoursPerDay / time.Hour,
		h: d % (HoursPerDay * time.Hour) / time.Hour,
		m: d % time.Hour / time.Minute,
		s: d % time.Minute / time.Second,
	}
}

// String 实现 fmt.Stringer
func (t TimeDuration) String() string {
	var s []string
	if 0 != t.d {
		s = append(s, fmt.Sprintf(`%d 天`, t.d))
	}
	if 0 != t.h {
		s = append(s, fmt.Sprintf(`%d 小时`, t.h))
	}
	if 0 != t.m {
		s = append(s, fmt.Sprintf(`%d 分钟`, t.m))
	}
	if 0 != t.s {
		s = append(s, fmt.Sprintf(`%d 秒`, t.s))
	}
	return strings.Join(s, ` `)
}
